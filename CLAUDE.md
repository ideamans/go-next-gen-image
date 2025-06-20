# github.com/ideamans/go-next-gen-image

golang で従来の Web フォーマット JPEG・PNG・GIF を、次世代画像フォーマット WebP と AVIF にベストプラクティスに従い変換するためのライブラリです。

## Windows CI の注意事項

Windows環境でのCIではChocolateyにvipsパッケージが存在しないため、代替手段を使用する必要があります：
1. vcpkgを使用してlibvipsをインストール
2. または、公式のWindowsバイナリを直接ダウンロード
3. または、Windowsテストをスキップ

# 仕様ライブラリ

- libvips

# 使い方

```go
converter := nextgenimage.NewConverter(nextgenimage.ConverterConfig{})

err := converter.ToWebP("input.jpg", "output.webp")
if err != nil {
    log.Fatal(err)
}

err := converter.ToAVIF("input.png", "output.avif")
if err != nil {
    log.Fatal(err)
}
```

# 変換ルール

## 共通

- Error は、FormatError(データに起因するエラー)と Error(その他のエラー)に分けて判定可能にする。
- ToWebP と ToAVIF は、変換後にファイルサイズが大きくなる場合は FormatError とする。

## JPEG to WebP

- lossy 変換
- Config の JPEGToWebP.Quality で品質を指定可能(デフォルト 80)
- EXIF でオリエンテーションが指定されている場合は先にデータを回転
- メタデータは ICC のみ維持

## PNG to WebP

- lossless 変換
- Config の PNGToWebP で TryNearLossless を指定可能(デフォルト false)
  - true の場合、NearLossless 変換も試し、サイズの小さい方を採用
- メタデータは ICC のみ維持

## GIF to WebP

- 各フレームは lossless 変換
- 時間やループなどの属性を維持して WebP アニメーションに変換

## WebP to AVIF

- Config の JPEGToAVIF.CQ で CQ を指定可能(デフォルト 25)

## PNG to AVIF

- lossless 変換

# テスト

- @testdata テストデータ集
- @testdata/index.json に目録があるので、それを元にできるだけ多くバリエーションについて ToWebP と ToAVIF をテストし、可否・サイズの削減・画角・PSNR による劣化の有無(閾値 40)を確認する。GIF についてはアニメーションの属性も判定する。
- その他異常系もカバーする

# lint

- golangci-lint をパスすること

# README

- README.md (英語)
- README.ja.md (日本語)

# CI

- GitHub Actions で go 1.22, 1.23 \* windows, linux, macos のマトリクスをテスト
- lint (linux \* go 1.23)をテスト
- coverage (linux \* go 1.23)をテスト

# LICENSE

- MIT

# ファイルマップ

- .github
- .gitignore
- README.md
- README.ja.md
- webp.go
- jpegtowebp_test.go
- pngtowebp_test.go
- gift_webp_test.go
- avif.go
- jpegtoavif_test.go
- pngtoavif_test.go
- converter.go
- converter_test.go
