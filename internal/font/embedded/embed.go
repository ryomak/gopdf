package embedded

import _ "embed"

// KoruriRegular は埋め込まれたKoruri Regularフォント
//
// Koruri（小瑠璃）は、M+ FONTSとOpen Sansを合成した日本語TrueTypeフォントです。
// 英数字部分がOpen Sansベースで美しく、日本語部分がM+ FONTSベースで読みやすいフォントです。
//
// ライセンス: Apache License 2.0
// 著作権: Koruri Project
// 構成: M+ FONTS + Open Sans
// サイズ: 約1.8MB
//
//go:embed Koruri-Regular.ttf
var KoruriRegular []byte

// License はKoruriフォントのライセンステキスト
//
//go:embed LICENSE.txt
var License string
