package config

func DefaultFlavor() string {
	return "debian"
}

func Flavor(flavor string) (string, bool) {
	switch flavor {
	case "debian":
		return flavor, true
	case "alpine":
		return flavor, true
	case "":
		return DefaultFlavor(), true
	default:
		return "", false
	}
}
