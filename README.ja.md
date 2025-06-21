# go-next-gen-image

従来のWeb画像フォーマット（JPEG、PNG、GIF）を次世代フォーマット（WebP、AVIF）にベストプラクティスに従って変換するGoライブラリです。

[![Go Reference](https://pkg.go.dev/badge/github.com/ideamans/go-next-gen-image.svg)](https://pkg.go.dev/github.com/ideamans/go-next-gen-image)
[![CI](https://github.com/ideamans/go-next-gen-image/actions/workflows/ci.yml/badge.svg)](https://github.com/ideamans/go-next-gen-image/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/ideamans/go-next-gen-image)](https://goreportcard.com/report/github.com/ideamans/go-next-gen-image)
[![codecov](https://codecov.io/gh/ideamans/go-next-gen-image/branch/main/graph/badge.svg)](https://codecov.io/gh/ideamans/go-next-gen-image)

## 機能

- JPEG/PNG/GIF画像をWebPフォーマットに変換
- JPEG/PNG画像をAVIFフォーマットに変換
- サイズ削減チェック付きの自動画像最適化
- 重要なメタデータ（ICCプロファイル）の保持
- アニメーションGIFからWebPアニメーションへの変換をサポート
- 品質設定のカスタマイズ可能
- スレッドセーフな並行変換

## インストール

```bash
go get github.com/ideamans/go-next-gen-image
```

### 前提条件

このライブラリはlibvipsがシステムにインストールされている必要があります：

**Ubuntu/Debian:**
```bash
sudo apt-get install libvips-dev
```

**macOS:**
```bash
brew install vips
```

**Windows:**
```bash
choco install vips
```

## 使い方

```go
package main

import (
    "log"
    nextgenimage "github.com/ideamans/go-next-gen-image"
)

func main() {
    // デフォルト設定でコンバーターを作成
    converter := nextgenimage.NewConverter(nextgenimage.ConverterConfig{})

    // JPEGをWebPに変換
    err := converter.ToWebP("input.jpg", "output.webp")
    if err != nil {
        log.Fatal(err)
    }

    // PNGをAVIFに変換
    err = converter.ToAVIF("input.png", "output.avif")
    if err != nil {
        log.Fatal(err)
    }
}
```

### 設定

```go
config := nextgenimage.ConverterConfig{
    JPEGToWebP: struct {
        Quality int
    }{
        Quality: 85, // デフォルト: 80
    },
    PNGToWebP: struct {
        TryNearLossless bool
    }{
        TryNearLossless: true, // デフォルト: false
    },
    JPEGToAVIF: struct {
        CQ int
    }{
        CQ: 20, // デフォルト: 25
    },
}

converter := nextgenimage.NewConverter(config)
```

## 変換ルール

### JPEG to WebP
- 損失圧縮
- 品質設定可能（デフォルト: 80）
- EXIFオリエンテーションに基づく自動回転
- EXIFとXMPメタデータを削除し、ICCプロファイルのみ保持

### PNG to WebP
- デフォルトで無損失圧縮
- より良い圧縮のためのオプションのニアロスレスモード
- EXIFとXMPメタデータを削除し、ICCプロファイルのみ保持
- アルファチャンネルのサポート

### GIF to WebP
- 無損失フレーム変換
- アニメーションの保持（タイミング、ループ）
- 全フレームをWebPアニメーションに変換

### JPEG to AVIF
- CQ（一定品質）モードでの損失圧縮
- CQ値の設定可能（デフォルト: 25、低いほど高品質）
- EXIFオリエンテーションに基づく自動回転
- EXIFとXMPメタデータを削除し、ICCプロファイルのみ保持

### PNG to AVIF
- 無損失圧縮
- EXIFとXMPメタデータを削除し、ICCプロファイルのみ保持
- アルファチャンネルのサポート

### GIF to AVIF
- サポートされていません（FormatErrorを返します）

## エラーハンドリング

ライブラリはデータ関連のエラーとシステムエラーを区別します：

```go
err := converter.ToWebP("input.jpg", "output.webp")
if err != nil {
    var formatErr *nextgenimage.FormatError
    if errors.As(err, &formatErr) {
        // フォーマット固有のエラーの処理（例：サポートされていないフォーマット、サイズ増加）
        log.Printf("フォーマットエラー: %v", err)
    } else {
        // システムエラーの処理（例：ファイルが見つからない、権限がない）
        log.Printf("システムエラー: %v", err)
    }
}
```

## パフォーマンス

- 効率的な画像処理のためにlibvipsを使用
- 並行変換をサポート（スレッドセーフ）
- 自動的に出力サイズをチェックし、変換後の画像が元より大きい場合はFormatErrorを返します

## テスト

全てのテストを実行：
```bash
go test ./...
```

カバレッジ付きでテストを実行：
```bash
go test -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## ライセンス

MITライセンス - 詳細は[LICENSE](LICENSE)ファイルを参照してください

## コントリビューション

1. リポジトリをフォーク
2. フィーチャーブランチを作成（`git checkout -b feature/amazing-feature`）
3. 変更をコミット（`git commit -m 'Add some amazing feature'`）
4. ブランチにプッシュ（`git push origin feature/amazing-feature`）
5. プルリクエストを開く

PRを送信する前に、適切にテストを更新し、全てのテストがパスすることを確認してください。