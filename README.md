# gored

gored は CLI ベースの Redmine 新規チケット作成ツールです。

## Description

* Redmine に新規チケットを登録します。
* チケットを登録する際、エディタを起動して登録内容を適宜編集できます。
* このとき、clipboard の内容がエディタの中に自動挿入されます。
* チケット登録に成功するとそのチケットのタイトルと URL を clipboard に追加します。

## Motivation

* 社内システムに関するイシューはイシューを立てた人はチケット登録行わず、イシューの内容をいったん ML に流すことになっている。
* ML に流れたイシューは、別の担当者が拾ってメールの内容を Redmine にコピペ登録した後にその URL を報告することになっている。
* 担当者 is 俺。
* いちいち Redmine にログインするのめんどい。ブラウザ開くのすらめんどい。コマンドラインベースでやりてぇ。

というわけで、コマンドラインから気楽にメールの内容をイシューとして新規登録し、その URL を返してくれるこのツールが生まれた。おぎゃー。

そもそもイシューを立てた人がチケット登録すればいいのでは？という問題提起ができないのは大人の事情なんですね :)

## Installation

[Releases ページ](https://github.com/yuta-masano/gored/releases)からダウンロードしてください。

あるいは、`go get` でも可能かもしれませんが、ライブラリパッケージは glide で vendoring しています。

```
$ go get github.com/yuta-masano/gored
```

## Usage

1. $HOME/.config/gored/config.yml または $HOME/.config/gored/config.json で以下を定義する。以下は config.yml の例。

   ```
   Endpoint: 'https://redmine.example.com'
   Apikey: アクセスキー
   Projects:
     1: 任意のプロジェクト名
     2: 任意のプロジェクト名
     ...
   Template: |-
     ### Single Line Subject ###
     h1. メール

     <pre>
     {{ .Clipboard }}
     </pre>
   ```

   `Projects` は各プロジェクトの project_id を key とし、project_id を同定するための任意のエイリアスを value とした連想配列です。
   各プロジェクトの project_id は例えば以下の URL から取得できます。
   ```
   https://redmine.example.com/projects.json
   ```

   `Template` はエディタで開くイシュー登録内容のテンプレートです。  
   一行目はイシューの「題名」として解釈され、二行目以降がイシューの「説明」と解釈されます。  
   `{{ .Clipboard }}` がクリップボードの内容に置き換えられます。

2. メールの内容を clipboard に登録しておく。
3. そのまま以下を実行。

   ```
   $ gored -t 'バグ' -p 'Normal' 任意のプロジェクト名
   ```

4. チケット登録に成功すると、以下のようなそのチケットのタイトルと URL が clipboard に追加される。

   ```
   [バグ #1234: ユーザ情報更新時に確認ポップアップが表示されない -  XXXX_プロジェクト - Redmine]
   https://redmine.example.com/issues/1234
   ```

### Option

```
$ gored --help
gored creates a new issue on Redmine using your clipboard text,
sends the added issue page's title and URL into your clipboard.

Usage:
  gored project_alias [flags]

Flags:
  -p, --priority string   choose Low, Normal, High (default "Normal")
  -t, --tracker string    choose 情報更新, バグ, 機能, サポート (default "バグ")
  -v, --version           show program's version number and exit
```

## License

The MIT License (MIT)

## Thanks to

* github.com/atotto/clipboard
* github.com/spf13/cobra
* github.com/mattn/go-redmine

## Author

[Yuta MASANO](https://github.com/yuta-masano)

## Development

### セットアップ

```
$ # 1. リポジトリを取得。
$ go get -v -u -d github.com/yuta-masano/gored

$ # 2. リポジトリディレクトリに移動。
$ cd $GOPATH/src/github.com/yuta-masano/gored

$ # 3. 開発ツールと vendor パッケージを取得。
$ make deps-install

$ # 4. その他のターゲットは help をどうぞ。
$ make help
USAGE: make [target]

TARGETS:
help           show help
...
```

### リリースフロー

see: [yuta-masano/dp#リリースフロー](https://github.com/yuta-masano/dp#%E3%83%AA%E3%83%AA%E3%83%BC%E3%82%B9%E3%83%95%E3%83%AD%E3%83%BC)
