package config

const (
	UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.97 Safari/537.36"
	MaxWorkers = 8
)

var Headers = map[string]string{
	"User-Agent": UserAgent,
}
