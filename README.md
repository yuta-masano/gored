# gored

gored は CLI ベースの Redmine 新規チケット作成ツールです。

## Description

* Redmine に clipboard の内容 を「説明」として新規チケット登録します。
* 引数でプロジェクト ID を指定します。
* オプションで「トラッカー」、「題名」、「優先度」が指定できます。
* チケット登録に成功するとそのチケットと URL を stdout に出力します。

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

## Usage

1. 先にメールの内容を clipboard に登録しておく。
2. そのまま以下を実行。
   ```
   $ gored -t 'バグ' -s 'ユーザ情報更新時に確認ポップアップが表示されない' -p 'normal' project_id
   ```
3. チケット登録に成功するとそのチケットのタイトルと URL が出力される。
   ```
   [バグ #1234: ユーザ情報更新時に確認ポップアップが表示されない - xxxx XXXX プロジェクト - Redmine]
   https://redmine.example.com/issues/1234
   ```

### Option

```

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
