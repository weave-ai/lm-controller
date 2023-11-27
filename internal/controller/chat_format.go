package controller

func getChatFormatFromModelFamily(family string) string {
	switch family {
	case "llama":
		return "llama-2"
	case "llama-2":
		return "llama-2"
	case "alpaca":
		return "alpaca"
	case "vicuna":
		return "vicuna"
	case "oasst-llama":
		return "oasst_llama"
	case "oasst_llama":
		return "oasst_llama"
	case "baichuan-2":
		return "baichuan-2"
	case "baichuan":
		return "baichuan"
	case "openbuddy":
		return "openbuddy"
	case "redpajama":
		return "redpajama-incite"
	case "redpajama-incite":
		return "redpajama-incite"
	case "snoozy":
		return "snoozy"
	case "phind":
		return "phind"
	case "intel":
		return "intel"
	case "open-orca":
		return "open-orca"
	case "mistrallite":
		return "mistrallite"
	case "zephyr":
		return "zephyr"
	case "chatml":
		return "chatml"
	case "openchat":
		return "openchat"
	}

	return "llama-2"
}
