# gored

gored は CLI ベースの Redmine 新規チケット作成ツールです。

## Description

* Redmine に新規チケットを登録します。
* チケットの「題名」、「説明」はコマンド実行時に起動するエディタで編集可能です。
* このとき、clipboard の内容が自動的に「説明」として扱われます。

* チケット登録に成功するとそのチケットのタイトルと URL を clipboard に追加します。

* 引数で「プロジェクト識別子」を指定します。
* オプションで「トラッカー」、「優先度」が指定できます。

## Motivation

* 社内システムに関するイシューはイシューを立てた人が直接チケット登録行わず、イシューの内容をいったん ML に流すことになっている。
* ML に流れたイシューは別の担当者が拾ってメールの内容を Redmine にコピペすることになっている。
* 担当者 is 俺。
* いちいち Redmine にログインするのめんどい。
* そもそもイシューを立てた人がチケット登録すればいいのでは？という問題提起ができないのは大人の事情なんですね :)

## Installation

```
$ go get github.com/yuta-masano/gored
```

または、[Release ページ](https://github.com/yuta-masano/gored/releases)からどうぞ。

## Usage

1. $HOME/.config/gored/config.yml または $HOME/.config/gored/config.json で以下を定義する。以下は config.yml の例。
   ```
   Endpoint: 'http://redmine.example.com'
   Apikey: アクセスキー
   Projects:
     1: プロジェクト識別子
     2: プロジェクト識別子
     ...
   ```
2. 先にメールの内容を clipboard に登録しておく。
3. そのまま以下を実行。
   ```
   $ gored -t 'バグ' -p 'normal' project_identifier
   ```
4. チケット登録に成功すると、以下の通りそのチケットのタイトルと URL が clipboard に追加される。
   ```
   [バグ #1234: ユーザ情報更新時に確認ポップアップが表示されない -  XXXX_プロジェクト - Redmine]
   https://redmine.example.com/issues/1234
   ```

### Option

```
$ gored --help
gored adds a new issue using your clipboard text,
returns the added issue pages's title and URL.

Usage:
  gored project_id [flags]

Flags:
  -p, --priority string   choose Low, Normal, High (default "Normal")
  -t, --tracker string    choose 情報更新, バグ, 機能, サポート (default "バグ")
```

## License

The MIT License (MIT)

## Author

[Yuta MASANO](https://github.com/yuta-masano)

## Development

### セットアップ

```
$ # 1. リポジトリを取得。
$ go get github.com/yuta-masano/gored

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
