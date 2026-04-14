package captcha

type visualItem struct {
	Key      string
	Label    string
	Icon     string
	Accent   string
	Category string
}

type categorySet struct {
	Key         string
	Label       string
	Prompt      string
	Accent      string
	Items       []visualItem
}

var categoryCatalog = []categorySet{
	{
		Key:    "food",
		Label:  "еда",
		Prompt: "можно съесть",
		Accent: "#f59e0b",
		Items: []visualItem{
			{Key: "apple", Label: "яблоко", Icon: "🍎", Accent: "#f97316", Category: "food"},
			{Key: "bread", Label: "хлеб", Icon: "🍞", Accent: "#f59e0b", Category: "food"},
			{Key: "pear", Label: "груша", Icon: "🍐", Accent: "#84cc16", Category: "food"},
			{Key: "cake", Label: "торт", Icon: "🍰", Accent: "#fb7185", Category: "food"},
		},
	},
	{
		Key:    "transport",
		Label:  "транспорт",
		Prompt: "ездит или летает",
		Accent: "#0ea5e9",
		Items: []visualItem{
			{Key: "car", Label: "машина", Icon: "🚗", Accent: "#38bdf8", Category: "transport"},
			{Key: "bus", Label: "автобус", Icon: "🚌", Accent: "#f97316", Category: "transport"},
			{Key: "train", Label: "поезд", Icon: "🚆", Accent: "#475569", Category: "transport"},
			{Key: "plane", Label: "самолет", Icon: "✈️", Accent: "#60a5fa", Category: "transport"},
		},
	},
	{
		Key:    "animals",
		Label:  "животные",
		Prompt: "живое",
		Accent: "#22c55e",
		Items: []visualItem{
			{Key: "fox", Label: "лиса", Icon: "🦊", Accent: "#f97316", Category: "animals"},
			{Key: "cat", Label: "кошка", Icon: "🐱", Accent: "#f59e0b", Category: "animals"},
			{Key: "whale", Label: "кит", Icon: "🐋", Accent: "#0ea5e9", Category: "animals"},
			{Key: "duck", Label: "утка", Icon: "🦆", Accent: "#84cc16", Category: "animals"},
		},
	},
	{
		Key:    "clothes",
		Label:  "одежда",
		Prompt: "можно надеть",
		Accent: "#a855f7",
		Items: []visualItem{
			{Key: "hat", Label: "шапка", Icon: "🧢", Accent: "#3b82f6", Category: "clothes"},
			{Key: "boot", Label: "ботинок", Icon: "🥾", Accent: "#92400e", Category: "clothes"},
			{Key: "shirt", Label: "рубашка", Icon: "👕", Accent: "#38bdf8", Category: "clothes"},
			{Key: "scarf", Label: "шарф", Icon: "🧣", Accent: "#ef4444", Category: "clothes"},
		},
	},
	{
		Key:    "furniture",
		Label:  "мебель",
		Prompt: "обычно стоит дома",
		Accent: "#14b8a6",
		Items: []visualItem{
			{Key: "chair", Label: "стул", Icon: "🪑", Accent: "#8b5cf6", Category: "furniture"},
			{Key: "lamp", Label: "лампа", Icon: "💡", Accent: "#facc15", Category: "furniture"},
			{Key: "bed", Label: "кровать", Icon: "🛏️", Accent: "#38bdf8", Category: "furniture"},
			{Key: "door", Label: "дверь", Icon: "🚪", Accent: "#92400e", Category: "furniture"},
		},
	},
}

type swipeScene struct {
	Title   string
	Body    string
	Icon    string
	Accent  string
	Possible bool
}

var possibleScenes = []swipeScene{
	{Title: "Проверка сцены", Body: "Кошка спит на подоконнике.", Icon: "🐱", Accent: "#f59e0b", Possible: true},
	{Title: "Проверка сцены", Body: "Поезд едет по рельсам.", Icon: "🚆", Accent: "#475569", Possible: true},
	{Title: "Проверка сцены", Body: "Утка плавает в пруду.", Icon: "🦆", Accent: "#22c55e", Possible: true},
	{Title: "Проверка сцены", Body: "Ребенок рисует мелом.", Icon: "🖍️", Accent: "#ef4444", Possible: true},
	{Title: "Проверка сцены", Body: "Повар режет хлеб.", Icon: "🍞", Accent: "#f97316", Possible: true},
}

var impossibleScenes = []swipeScene{
	{Title: "Проверка сцены", Body: "Кит варит суп на Луне.", Icon: "🐋", Accent: "#0ea5e9", Possible: false},
	{Title: "Проверка сцены", Body: "Облако едет на велосипеде.", Icon: "☁️", Accent: "#38bdf8", Possible: false},
	{Title: "Проверка сцены", Body: "Стул пьет лимонад.", Icon: "🪑", Accent: "#14b8a6", Possible: false},
	{Title: "Проверка сцены", Body: "Ботинок поет в микрофон.", Icon: "🥾", Accent: "#92400e", Possible: false},
	{Title: "Проверка сцены", Body: "Торт сам открывает дверь.", Icon: "🍰", Accent: "#fb7185", Possible: false},
}

var russianUppercase = []string{"А", "Б", "В", "Г", "Д", "Е", "Ж", "З", "И", "К", "Л", "М", "Н", "О", "П", "Р", "С", "Т", "У", "Ф", "Х", "Ц", "Ч", "Ш", "Щ", "Ы", "Э", "Ю", "Я"}
var russianLowercase = []string{"а", "б", "в", "г", "д", "е", "ж", "з", "и", "к", "л", "м", "н", "о", "п", "р", "с", "т", "у", "ф", "х", "ц", "ч", "ш", "щ", "ы", "э", "ю", "я"}
var foreignUppercase = []string{"A", "B", "C", "D", "F", "G", "Q", "R", "Ω", "Δ", "Σ", "N"}
var foreignLowercase = []string{"a", "e", "o", "p", "x", "y", "q", "m", "β", "δ"}

type basketRule struct {
	Key         string
	LeftLabel   string
	RightLabel  string
	LeftAccent  string
	RightAccent string
	Items       []basketItem
}

type basketItem struct {
	Key      string
	Label    string
	Icon     string
	LeftSide bool
}

var basketRules = []basketRule{
	{
		Key:         "living",
		LeftLabel:   "Живое",
		RightLabel:  "Не живое",
		LeftAccent:  "#22c55e",
		RightAccent: "#64748b",
		Items: []basketItem{
			{Key: "cat", Label: "кошка", Icon: "🐱", LeftSide: true},
			{Key: "fox", Label: "лиса", Icon: "🦊", LeftSide: true},
			{Key: "whale", Label: "кит", Icon: "🐋", LeftSide: true},
			{Key: "chair", Label: "стул", Icon: "🪑", LeftSide: false},
			{Key: "lamp", Label: "лампа", Icon: "💡", LeftSide: false},
			{Key: "train", Label: "поезд", Icon: "🚆", LeftSide: false},
		},
	},
	{
		Key:         "edible",
		LeftLabel:   "Съедобное",
		RightLabel:  "Не съедобное",
		LeftAccent:  "#f59e0b",
		RightAccent: "#8b5cf6",
		Items: []basketItem{
			{Key: "apple", Label: "яблоко", Icon: "🍎", LeftSide: true},
			{Key: "bread", Label: "хлеб", Icon: "🍞", LeftSide: true},
			{Key: "cake", Label: "торт", Icon: "🍰", LeftSide: true},
			{Key: "boot", Label: "ботинок", Icon: "🥾", LeftSide: false},
			{Key: "bus", Label: "автобус", Icon: "🚌", LeftSide: false},
			{Key: "door", Label: "дверь", Icon: "🚪", LeftSide: false},
		},
	},
	{
		Key:         "letters",
		LeftLabel:   "Русская буква",
		RightLabel:  "Не русская",
		LeftAccent:  "#0ea5e9",
		RightAccent: "#ef4444",
		Items: []basketItem{
			{Key: "ru_a", Label: "А", Icon: "", LeftSide: true},
			{Key: "ru_m", Label: "М", Icon: "", LeftSide: true},
			{Key: "ru_yu", Label: "Ю", Icon: "", LeftSide: true},
			{Key: "latin_q", Label: "Q", Icon: "", LeftSide: false},
			{Key: "latin_b", Label: "B", Icon: "", LeftSide: false},
			{Key: "greek_omega", Label: "Ω", Icon: "", LeftSide: false},
		},
	},
}

type trackPiece struct {
	Key    string
	Label  string
	Accent string
}

var trackPieces = []trackPiece{
	{Key: "red", Label: "●", Accent: "#ef4444"},
	{Key: "blue", Label: "▲", Accent: "#3b82f6"},
	{Key: "green", Label: "■", Accent: "#22c55e"},
	{Key: "amber", Label: "◆", Accent: "#f59e0b"},
}
