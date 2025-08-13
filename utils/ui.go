package utils

// GetFileIcon returns an appropriate icon for the file extension
func GetFileIcon(ext string) string {
	icons := map[string]string{
		// Programming languages
		".go":    "🐹",
		".py":    "🐍",
		".js":    "📄",
		".ts":    "📘",
		".jsx":   "⚛️",
		".tsx":   "⚛️",
		".rs":    "🦀",
		".java":  "☕",
		".c":     "📄",
		".cpp":   "📄",
		".cc":    "📄",
		".h":     "📄",
		".hpp":   "📄",
		".cs":    "🔷",
		".php":   "🐘",
		".rb":    "💎",
		".swift": "🍎",
		".kt":    "📱",
		".scala": "⚖️",

		// Documentation and markup
		".md":       "📝",
		".markdown": "📝",
		".txt":      "📄",
		".rst":      "📜",
		".tex":      "📰",

		// Configuration files
		".json":       "🔧",
		".yaml":       "⚙️",
		".yml":        "⚙️",
		".toml":       "⚙️",
		".ini":        "⚙️",
		".cfg":        "⚙️",
		".conf":       "⚙️",
		".env":        "🌿",
		".properties": "⚙️",

		// Data files
		".xml": "📋",
		".csv": "📊",
		".log": "📜",
		".sql": "🗄️",

		// Shell and scripts
		".sh":   "🐚",
		".bash": "🐚",
		".zsh":  "🐚",
		".fish": "🐠",
		".ps1":  "💻",
		".bat":  "💻",
		".cmd":  "💻",

		// Build and package files
		".dockerfile": "🐳",
		".makefile":   "🔨",
		".gradle":     "🐘",
		".pom":        "📦",
		".package":    "📦",

		// Web and frontend
		".html": "🌐",
		".htm":  "🌐",
		".css":  "🎨",
		".scss": "🎨",
		".sass": "🎨",
		".less": "🎨",

		// Images
		".png":  "🖼️",
		".jpg":  "🖼️",
		".jpeg": "🖼️",
		".gif":  "🖼️",
		".svg":  "🖼️",
		".ico":  "🖼️",

		// Archives
		".zip": "📦",
		".tar": "📦",
		".gz":  "📦",
		".rar": "📦",
		".7z":  "📦",

		// Executables
		".exe": "⚙️",
		".bin": "⚙️",
		".deb": "📦",
		".rpm": "📦",
		".msi": "📦",
	}

	if icon, exists := icons[ext]; exists {
		return icon
	}
	return "📄"
}
