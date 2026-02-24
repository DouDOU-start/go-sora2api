package pow

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"go-sora2api/internal/util"

	"golang.org/x/crypto/sha3"
)

const maxIteration = 500000

var (
	cores   = []int{8, 16, 24, 32}
	screens = []int{3000, 4000, 3120, 4160}

	scripts = []string{
		"https://cdn.oaistatic.com/_next/static/cXh69klOLzS0Gy2joLDRS/_ssgManifest.js?dpl=453ebaec0d44c2decab71692e1bfe39be35a24b3",
	}
	dpl = []string{
		"prod-f501fe933b3edf57aea882da888e1a544df99840",
	}

	// 注意: 分隔符是 U+2212 (MINUS SIGN)，不是 U+002D (HYPHEN-MINUS)
	navigatorKeys = []string{
		"registerProtocolHandler\u2212function registerProtocolHandler() { [native code] }",
		"storage\u2212[object StorageManager]",
		"locks\u2212[object LockManager]",
		"appCodeName\u2212Mozilla",
		"permissions\u2212[object Permissions]",
		"webdriver\u2212false",
		"vendor\u2212Google Inc.",
		"mediaDevices\u2212[object MediaDevices]",
		"cookieEnabled\u2212true",
		"product\u2212Gecko",
		"productSub\u221220030107",
		"hardwareConcurrency\u221232",
		"onLine\u2212true",
	}
	documentKeys = []string{"_reactListeningo743lnnpvdg", "location"}
	windowKeys   = []string{
		"0", "window", "self", "document", "name", "location",
		"navigator", "screen", "innerWidth", "innerHeight",
		"localStorage", "sessionStorage", "crypto", "performance",
		"fetch", "setTimeout", "setInterval", "console",
	}
)

// getParseTime 生成 EST 时区时间字符串
func getParseTime() string {
	loc := time.FixedZone("EST", -5*3600)
	now := time.Now().In(loc)
	return now.Format("Mon Jan 02 2006 15:04:05") + " GMT-0500 (Eastern Standard Time)"
}

// getConfig 构造 18 元素的浏览器指纹数组
func getConfig(userAgent string) []interface{} {
	perfCounter := float64(time.Now().UnixNano()%1e12) / 1e6
	timeMs := float64(time.Now().UnixMilli())

	return []interface{}{
		screens[rand.Intn(len(screens))],
		getParseTime(),
		4294705152,
		0,
		userAgent,
		scripts[rand.Intn(len(scripts))],
		dpl[rand.Intn(len(dpl))],
		"en-US",
		"en-US,es-US,en,es",
		0,
		navigatorKeys[rand.Intn(len(navigatorKeys))],
		documentKeys[rand.Intn(len(documentKeys))],
		windowKeys[rand.Intn(len(windowKeys))],
		perfCounter,
		util.GenerateUUID(),
		"",
		cores[rand.Intn(len(cores))],
		timeMs - perfCounter,
	}
}

func compactJSON(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}

// solve 执行 PoW 计算：SHA3-512 哈希碰撞
func solve(seed, difficulty string, configList []interface{}) (string, bool) {
	diffBytes, _ := hex.DecodeString(difficulty)
	diffLen := len(diffBytes)
	seedBytes := []byte(seed)

	part1JSON := compactJSON(configList[:3])
	staticPart1 := append(part1JSON[:len(part1JSON)-1], ',')

	part2JSON := compactJSON(configList[4:9])
	staticPart2 := make([]byte, 0, len(part2JSON)+2)
	staticPart2 = append(staticPart2, ',')
	staticPart2 = append(staticPart2, part2JSON[1:len(part2JSON)-1]...)
	staticPart2 = append(staticPart2, ',')

	part3JSON := compactJSON(configList[10:])
	staticPart3 := make([]byte, 0, len(part3JSON)+1)
	staticPart3 = append(staticPart3, ',')
	staticPart3 = append(staticPart3, part3JSON[1:]...)

	for i := 0; i < maxIteration; i++ {
		dynamicI := []byte(strconv.Itoa(i))
		dynamicJ := []byte(strconv.Itoa(i >> 1))

		totalLen := len(staticPart1) + len(dynamicI) + len(staticPart2) + len(dynamicJ) + len(staticPart3)
		finalJSON := make([]byte, 0, totalLen)
		finalJSON = append(finalJSON, staticPart1...)
		finalJSON = append(finalJSON, dynamicI...)
		finalJSON = append(finalJSON, staticPart2...)
		finalJSON = append(finalJSON, dynamicJ...)
		finalJSON = append(finalJSON, staticPart3...)

		b64 := make([]byte, base64.StdEncoding.EncodedLen(len(finalJSON)))
		base64.StdEncoding.Encode(b64, finalJSON)

		h := sha3.New512()
		h.Write(seedBytes)
		h.Write(b64)
		hash := h.Sum(nil)

		if bytesLessOrEqual(hash[:diffLen], diffBytes) {
			return string(b64), true
		}
	}

	errorToken := "wQ8Lk5FbGpA2NcR9dShT6gYjU7VxZ4D" +
		base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(`"%s"`, seed)))
	return errorToken, false
}

func bytesLessOrEqual(a, b []byte) bool {
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] < b[i] {
			return true
		}
		if a[i] > b[i] {
			return false
		}
	}
	return true
}

func mustJSONStr(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

// GetToken 生成初始 PoW token
func GetToken(userAgent string) string {
	configList := getConfig(userAgent)
	seed := strconv.FormatFloat(rand.Float64(), 'f', -1, 64)
	solution, _ := solve(seed, "0fffff", configList)
	return "gAAAAAC" + solution
}

// BuildSentinelToken 从 sentinel/req 响应构建最终的 sentinel token
func BuildSentinelToken(flow, reqID, powToken string, resp map[string]interface{}, userAgent string) string {
	finalPowToken := powToken

	if proofofwork, ok := resp["proofofwork"].(map[string]interface{}); ok {
		if required, _ := proofofwork["required"].(bool); required {
			seed, _ := proofofwork["seed"].(string)
			difficulty, _ := proofofwork["difficulty"].(string)
			if seed != "" && difficulty != "" {
				configList := getConfig(userAgent)
				solution, success := solve(seed, difficulty, configList)
				finalPowToken = "gAAAAAB" + solution
				if !success {
					fmt.Println("[警告] PoW 计算失败，使用错误 token")
				}
			}
		}
	}

	if len(finalPowToken) < 2 || finalPowToken[len(finalPowToken)-2:] != "~S" {
		finalPowToken += "~S"
	}

	turnstileDX := ""
	if turnstile, ok := resp["turnstile"].(map[string]interface{}); ok {
		turnstileDX, _ = turnstile["dx"].(string)
	}

	tokenStr, _ := resp["token"].(string)

	result := fmt.Sprintf(`{"p":%s,"t":%s,"c":%s,"id":%s,"flow":%s}`,
		mustJSONStr(finalPowToken),
		mustJSONStr(turnstileDX),
		mustJSONStr(tokenStr),
		mustJSONStr(reqID),
		mustJSONStr(flow),
	)
	return result
}
