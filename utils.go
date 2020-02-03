// 🔌 Fiber is an Express.js inspired web framework build on 🚀 Fasthttp.
// 📌 Please open an issue if you got suggestions or found a bug!
// 🖥 https://github.com/gofiber/fiber

// 🦸 Not all heroes wear capes, thank you to some amazing people
// 💖 @valyala, @dgrr, @erikdubbelboer, @savsgio, @julienschmidt

package fiber

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unsafe"
)

func getParams(path string) (params []string) {
	segments := strings.Split(path, "/")
	replacer := strings.NewReplacer(":", "", "?", "")
	for _, s := range segments {
		if s == "" {
			continue
		} else if s[0] == ':' {
			params = append(params, replacer.Replace(s))
		} else if s[0] == '*' {
			params = append(params, "*")
		}
	}
	return params
}

func getRegex(path string) (*regexp.Regexp, error) {
	pattern := "^"
	segments := strings.Split(path, "/")
	for _, s := range segments {
		if s[0] == ':' {
			if strings.Contains(s, "?") {
				pattern += "(?:/([^/]+?))?"
			} else {
				pattern += "/(?:([^/]+?))"
			}
		} else if s[0] == '*' {
			pattern += "/(.*)"
		} else {
			pattern += "/" + s
		}
	}
	pattern += "/?$"
	regex, err := regexp.Compile(pattern)
	return regex, err
}

func getFiles(root string) (files []string, isDir bool, err error) {
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		} else {
			isDir = true
		}
		return err
	})
	return files, isDir, err
}

func getType(ext string) (mime string) {
	if ext[0] == '.' {
		ext = ext[1:]
	}
	mime = contentTypes[ext]
	if mime == "" {
		return contentTypeOctetStream
	}
	return mime
}

func getStatus(status int) (msg string) {
	return statusMessages[status]
}

func getString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func getBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(&s))
}
