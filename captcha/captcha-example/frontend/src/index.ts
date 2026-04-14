import './styles.scss'

class App {
    public static sendDataToServer(data: Uint8Array) {
        window.top.postMessage({type:'captcha:sendData',data:data})
    }

    public static handleServerData(data: ArrayBuffer) {
        console.log("Got server data", new Uint8Array(data))
    }

    public static run() {
        window.addEventListener("message", (e) => {
            if (e.data.type === "captcha:serverData") {
                this.handleServerData(e.data.data)
            }
        })

        document.getElementById("testSend").onclick = () => {
            App.sendDataToServer(Uint8Array.of(0b10000000, 0b00000000))
        }
        document.getElementById("result20").onclick = () => {
            App.sendDataToServer(Uint8Array.of(0b10000000, 0b00000001))
        }
        document.getElementById("result50").onclick = () => {
            App.sendDataToServer(Uint8Array.of(0b10000000, 0b00000010))
        }
        document.getElementById("result80").onclick = () => {
            App.sendDataToServer(Uint8Array.of(0b10000000, 0b00000011))
        }
    }
}

App.run()