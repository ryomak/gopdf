package content

import (
	"bytes"
	"fmt"
	"strconv"
)

// ToUnicodeCMap はCIDからUnicodeへのマッピング
type ToUnicodeCMap struct {
	charMap map[uint16]rune   // 個別のCIDマッピング (bfchar)
	ranges  []CIDRange        // 範囲マッピング (bfrange)
}

// GetCharMapSize は charMap のエントリ数を返す（デバッグ用）
func (cm *ToUnicodeCMap) GetCharMapSize() int {
	if cm == nil {
		return 0
	}
	return len(cm.charMap)
}

// GetRangesSize は ranges のエントリ数を返す（デバッグ用）
func (cm *ToUnicodeCMap) GetRangesSize() int {
	if cm == nil {
		return 0
	}
	return len(cm.ranges)
}

// GetSampleMappings はサンプルマッピングを返す（デバッグ用）
func (cm *ToUnicodeCMap) GetSampleMappings(count int) map[uint16]rune {
	if cm == nil || count <= 0 {
		return nil
	}
	samples := make(map[uint16]rune)
	i := 0
	for cid, r := range cm.charMap {
		if i >= count {
			break
		}
		samples[cid] = r
		i++
	}
	return samples
}

// CIDRange はCIDの範囲マッピング
type CIDRange struct {
	StartCID  uint16
	EndCID    uint16
	StartChar rune
}

// Lookup はCIDをUnicodeに変換
func (cm *ToUnicodeCMap) Lookup(cid uint16) (rune, bool) {
	if cm == nil {
		return 0, false
	}

	// 1. charMapで検索
	if r, ok := cm.charMap[cid]; ok {
		return r, true
	}

	// 2. rangesで検索
	for _, rang := range cm.ranges {
		if cid >= rang.StartCID && cid <= rang.EndCID {
			offset := cid - rang.StartCID
			return rang.StartChar + rune(offset), true
		}
	}

	return 0, false
}

// LookupString はCIDバイト列をUnicode文字列に変換
func (cm *ToUnicodeCMap) LookupString(data []byte) string {
	if cm == nil || len(data) == 0 {
		return ""
	}

	var result []rune

	// 2バイトずつCIDとして読み取り（ビッグエンディアン）
	for i := 0; i < len(data); i += 2 {
		if i+1 >= len(data) {
			break
		}

		cid := uint16(data[i])<<8 | uint16(data[i+1])

		if r, ok := cm.Lookup(cid); ok {
			result = append(result, r)
		} else {
			// マッピングがない場合は元のCIDを使用（デバッグ用）
			// または置換文字を使用
			result = append(result, rune(cid))
		}
	}

	return string(result)
}

// ParseToUnicodeCMap はToUnicode CMapをパースする
func ParseToUnicodeCMap(data []byte) (*ToUnicodeCMap, error) {
	cmap := &ToUnicodeCMap{
		charMap: make(map[uint16]rune),
	}

	// beginbfchar/endbfchar セクションをパース
	charMaps, err := parseBFChar(data)
	if err != nil {
		return nil, fmt.Errorf("parse bfchar: %w", err)
	}
	cmap.charMap = charMaps

	// beginbfrange/endbfrange セクションをパース
	ranges, err := parseBFRange(data)
	if err != nil {
		return nil, fmt.Errorf("parse bfrange: %w", err)
	}
	cmap.ranges = ranges

	return cmap, nil
}

// parseBFChar は beginbfchar セクションをパース
func parseBFChar(data []byte) (map[uint16]rune, error) {
	result := make(map[uint16]rune)

	// "beginbfchar" を検索
	beginMarker := []byte("beginbfchar")
	endMarker := []byte("endbfchar")

	startIdx := 0
	for {
		idx := bytes.Index(data[startIdx:], beginMarker)
		if idx == -1 {
			break
		}
		startIdx += idx + len(beginMarker)

		// "endbfchar" までを抽出
		endIdx := bytes.Index(data[startIdx:], endMarker)
		if endIdx == -1 {
			break
		}

		section := data[startIdx : startIdx+endIdx]

		// <XXXX> <YYYY> のペアを抽出
		pairs := extractHexPairs(section)
		for _, pair := range pairs {
			if len(pair) == 2 {
				cid := parseHex(pair[0])
				unicode := parseHex(pair[1])
				if cid < 0x10000 && unicode >= 0 {
					result[uint16(cid)] = rune(unicode)
				}
			}
		}

		startIdx += endIdx + len(endMarker)
	}

	return result, nil
}

// parseBFRange は beginbfrange セクションをパース
func parseBFRange(data []byte) ([]CIDRange, error) {
	var result []CIDRange

	// "beginbfrange" を検索
	beginMarker := []byte("beginbfrange")
	endMarker := []byte("endbfrange")

	startIdx := 0
	for {
		idx := bytes.Index(data[startIdx:], beginMarker)
		if idx == -1 {
			break
		}
		startIdx += idx + len(beginMarker)

		// "endbfrange" までを抽出
		endIdx := bytes.Index(data[startIdx:], endMarker)
		if endIdx == -1 {
			break
		}

		section := data[startIdx : startIdx+endIdx]

		// <XXXX> <YYYY> <ZZZZ> のトリプルを抽出
		triples := extractHexTriples(section)
		for _, triple := range triples {
			if len(triple) == 3 {
				startCID := parseHex(triple[0])
				endCID := parseHex(triple[1])
				startChar := parseHex(triple[2])

				if startCID < 0x10000 && endCID < 0x10000 && startChar >= 0 {
					result = append(result, CIDRange{
						StartCID:  uint16(startCID),
						EndCID:    uint16(endCID),
						StartChar: rune(startChar),
					})
				}
			}
		}

		startIdx += endIdx + len(endMarker)
	}

	return result, nil
}

// extractHexPairs は <XXXX> <YYYY> 形式のペアを抽出
func extractHexPairs(data []byte) [][]string {
	var result [][]string
	var current []string

	i := 0
	for i < len(data) {
		// '<' を探す
		if data[i] == '<' {
			j := i + 1
			// '>' を探す
			for j < len(data) && data[j] != '>' {
				j++
			}
			if j < len(data) {
				hexStr := string(data[i+1 : j])
				current = append(current, hexStr)

				// ペアが揃ったら追加
				if len(current) == 2 {
					result = append(result, current)
					current = nil
				}

				i = j + 1
				continue
			}
		}
		i++
	}

	return result
}

// extractHexTriples は <XXXX> <YYYY> <ZZZZ> 形式のトリプルを抽出
func extractHexTriples(data []byte) [][]string {
	var result [][]string
	var current []string

	i := 0
	for i < len(data) {
		// '<' を探す
		if data[i] == '<' {
			j := i + 1
			// '>' を探す
			for j < len(data) && data[j] != '>' {
				j++
			}
			if j < len(data) {
				hexStr := string(data[i+1 : j])
				current = append(current, hexStr)

				// トリプルが揃ったら追加
				if len(current) == 3 {
					result = append(result, current)
					current = nil
				}

				i = j + 1
				continue
			}
		}
		i++
	}

	return result
}

// parseHex は16進数文字列を整数に変換
func parseHex(hexStr string) int {
	val, err := strconv.ParseInt(hexStr, 16, 64)
	if err != nil {
		return -1
	}
	return int(val)
}
