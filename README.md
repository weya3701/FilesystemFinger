# FilesystemFinger

`fsfinger` 可為檔案與目錄樹建立具確定性的 SHA-256 指紋。當相對路徑、物件類型、
符號連結目標、空目錄或檔案內容發生變更時，目錄指紋也會隨之改變；但僅移動掃描
根目錄本身的位置，並不會改變指紋。

權限與修改時間等中繼資料會包含在 JSON 清單中供檢查，但不會參與根雜湊值的計算。

## 建置與使用

```sh
go build -o fsfinger ./cmd/fsfinger

./fsfinger scan /path/to/directory
./fsfinger scan --hash-only /path/to/directory
./fsfinger scan /path/to/directory > manifest.json
./fsfinger verify --manifest manifest.json /path/to/directory
```

請將清單寫入受掃描目錄以外的位置，或忽略其路徑，以免輸出檔案成為下次掃描內容的
一部分。

## 作為 Go 模組使用

將此模組加入其他 Go 專案：

```sh
go get github.com/ccxn/filesystemfinger
```

呼叫 `fingerprint.Hash` 可掃描檔案、目錄樹或符號連結，並僅回傳雜湊值；其效果等同
於 `fsfinger scan --hash-only`：

```go
package main

import (
	"fmt"
	"log"

	"github.com/ccxn/filesystemfinger/fingerprint"
)

func main() {
	hash, err := fingerprint.Hash("/path/to/directory", fingerprint.Options{
		UseDefaults: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(hash)
}
```

如需自訂忽略規則，可在 `fingerprint.Options` 中設定 `IgnorePatterns` 或
`IgnoreFile`。若需要可直接轉換為 JSON 的完整清單及個別項目，請改用
`fingerprint.Scan`。

## 忽略規則

內建規則會忽略 `.git/`、`.DS_Store`、`Thumbs.db`、`*.swp` 與 `*~`。可重複使用
`--ignore` 加入多個模式、透過 `--ignore-file` 從檔案載入規則，或使用
`--no-default-ignore` 停用預設規則。

所有平台上的模式一律使用 `/` 作為分隔符號。`*` 會比對單一路徑區段內的字元，
`**` 可跨越多層目錄，`?` 會比對一個字元，結尾的 `/` 僅比對目錄，而開頭的 `!`
則會取消先前的規則。

## 穩定格式

名稱與符號連結目標會正規化為 Unicode NFC，路徑則使用 `/`。目錄中的子項目會依
正規化後的 UTF-8 位元組排序。標準記錄會加上長度前綴、採用大端序長度，並以
`filesystem-fingerprint-v1` 格式識別碼進行域分隔。檔案會完全依照原始位元組計算
雜湊，不會改寫行尾字元；掃描時也不會跟隨符號連結。
