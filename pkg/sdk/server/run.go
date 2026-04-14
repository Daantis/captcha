package server

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"os/signal"
	v1 "sdk/pkg/pb/v1"
	"strconv"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

const (
	defaultMinPort = "38000"
	defaultMaxPort = "40000"

	//	defaultHost       = "localhost"
	//	defaultTargetHost = "localhost:50051"
)

const (
	envHost    = "HOST"
	envMinPort = "MIN_PORT"
	envMaxPort = "MAX_PORT"

	envTargetHost  = "TARGET_HOST"
	envChallengeId = "CHALLENGE_ID"
)

func Run() {
	// start server
	// connect to remote server
	// send heartbeats
	// read responses
	slog.Info("starting server")

	host := mustGetEnv(envHost)
	targetHost := mustGetEnv(envTargetHost)

	challengeId := ChallengeId(mustGetEnv(envChallengeId))
	instanceId := InstanceId(uuid.NewString())

	// building captcher
	builder := GetCaptcherBuilder()
	captcher, err := builder(challengeId, instanceId)
	if err != nil {
		panic(err)
	}

	srv, err := newServer(captcher)
	if err != nil {
		panic(err)
	}

	// init signals
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// start grpc server on the free port
	// default 38000-40000
	minPort, err := strconv.Atoi(getEnvOrDefault(envMinPort, defaultMinPort))
	if err != nil {
		panic(errors.Wrap(err, "invalid min port"))
	}

	maxPort, err := strconv.Atoi(getEnvOrDefault(envMaxPort, defaultMaxPort))
	if err != nil {
		panic(errors.Wrap(err, "invalid max port"))
	}

	if maxPort < minPort {
		panic(errors.New("max port must be greater or equal than min port"))
	}

	var lis net.Listener
	for minPort <= maxPort {
		lis, err = net.Listen("tcp", fmt.Sprintf(":%d", minPort))
		if err == nil {
			break
		}

		minPort++
	}

	if lis == nil {
		panic(errors.New("no free port available"))
	}

	port := int32(minPort)

	// create connection for target host
	conn, err := grpc.NewClient(
		fmt.Sprintf("dns:///%s", targetHost),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	client := v1.NewBalancerServiceClient(conn)

	instanceStream, err := client.RegisterInstance(context.TODO())
	if err != nil {
		panic(err)
	}

	defer instanceStream.Send(&v1.RegisterInstanceRequest{
		EventType:   v1.RegisterInstanceRequest_STOPPED,
		InstanceId:  instanceId.String(),
		Host:        host,
		PortNumber:  port,
		ChallengeId: challengeId.String(),
		Timestamp:   time.Now().Unix(),
	})
	defer instanceStream.CloseSend()

	// start server
	startHeartbeat := make(chan struct{})

	hbCtx, cancel := context.WithCancel(context.Background())
	go func() {
		close(startHeartbeat)
		if err := srv.Serve(lis); err != nil {
			slog.Error(err.Error())
			cancel()
		}
	}()

	<-startHeartbeat
	err = instanceStream.Send(&v1.RegisterInstanceRequest{
		EventType:   v1.RegisterInstanceRequest_READY,
		InstanceId:  instanceId.String(),
		Host:        host,
		PortNumber:  port,
		Timestamp:   time.Now().Unix(),
		ChallengeId: challengeId.String(),
	})
	if err != nil {
		slog.Error("failed to send first heartbeat", slog.String("error", err.Error()))
	}

	go func() {
		// send heartbeats
		ticker := time.NewTicker(time.Second)
		for {
			select {
			case <-ticker.C:
				// todo(nth): why don't send live with instance id only?
				err = instanceStream.Send(&v1.RegisterInstanceRequest{
					EventType:   v1.RegisterInstanceRequest_READY,
					InstanceId:  instanceId.String(),
					Host:        host,
					PortNumber:  port,
					Timestamp:   time.Now().Unix(),
					ChallengeId: challengeId.String(),
				})
				if err != nil {
					slog.Error("failed to send heartbeat", slog.String("err", err.Error()))
				}
			case <-instanceStream.Context().Done():
				return

			case <-hbCtx.Done():
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case <-instanceStream.Context().Done():
				return
			case <-hbCtx.Done():
				return
			default:
			}

			in, err := instanceStream.Recv()
			if err == io.EOF {
				slog.Debug("received EOF")
				// no more messages
				return
			}

			if err != nil {
				slog.Error("received error from stream", slog.String("err", err.Error()))
				continue
			}

			switch in.Status {
			case v1.RegisterInstanceResponse_SUCCESS:
				// slog.Info("instance is ready")
				// everything is alright
				continue
			case v1.RegisterInstanceResponse_ERROR:
				slog.Error("failed to register instance", slog.String("error", err.Error()))
				continue
			default:
				slog.Error("unknown instance response status", slog.Int("status", int(in.Status)))
				continue
			}
		}
	}()

	<-signals
	if err = srv.Stop(); err != nil {
		slog.Error(err.Error())
	}
}

type grpcServer struct {
	server *grpc.Server
	lis    net.Listener
	closer incomingStreamCloser
}

func (s *grpcServer) Serve(lis net.Listener) error {
	s.lis = lis
	return s.server.Serve(lis)
}

func (s *grpcServer) Stop() error {
	err := s.closer.closeIncomingStream()
	if err != nil {
		slog.Error(err.Error())
	}

	if err := s.lis.Close(); err != nil {
		return err
	}

	shutdownTimeout := 10 * time.Minute
	done := make(chan struct{})

	go func() {
		s.server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		slog.Info("gracefully shut down")
	case <-time.After(shutdownTimeout):
		s.server.Stop()
		slog.Info("shut down by timeout")
	}
	return nil
}

func newServer(c Captcher) (*grpcServer, error) {
	ka := keepalive.ServerParameters{
		MaxConnectionAge:      10 * time.Minute,
		MaxConnectionAgeGrace: 10 * time.Minute,
		MaxConnectionIdle:     10 * time.Minute,
		Time:                  5 * time.Minute,
		Timeout:               20 * time.Second,
	}
	clsr := newCloser()

	srv := grpc.NewServer(
		grpc.KeepaliveParams(ka),
		grpc.MaxConcurrentStreams(1024),
		grpc.ConnectionTimeout(10*time.Second),
		grpc.ChainStreamInterceptor(streamCloseInterceptor(clsr)),
		grpc.ChainUnaryInterceptor(unaryCloseInterceptor(clsr)),
	)

	v1.RegisterCaptchaServiceServer(srv, mustBuildHandler(c, clsr))

	return &grpcServer{server: srv, closer: clsr}, nil
}

func mustGetEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("environment variable %s not set", key))
	}

	return val
}

func getEnvOrDefault(key string, defaultValue string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}

	return val
}
