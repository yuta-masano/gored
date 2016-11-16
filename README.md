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
   Endpoint: 'https://redmine.example.com'
   Apikey: アクセスキー
   Projects:
     1: プロジェクト alias
     2: プロジェクト alias
     ...
   ```

   ```Projects``` は各 プロジェクトの project_id を key とし、project_id を同定するための任意のエイリアスを value とした連想配列です。
   各プロジェクトの project_id は例えば以下で確認できます。
   ```
   $ psql -U postgres -At -d redmine -c "SELECT id || ': ' || identifier FROM projects ORDER BY id"
   1: xxx-prj
   2: yyy-prj
   3: zzz-prj
   ...
   ```

2. 先にメールの内容を clipboard に登録しておく。
3. そのまま以下を実行。

   ```
   $ gored -t 'バグ' -p 'Normal' project_alias
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

1. ローカルブランチを切る。  
   ブランチ名にはリリースするバージョン番号を含めること。
2. コミット コミット コミット ...  
   コミットログの一部は CHANGELOG に使いまわす想定。  
   以下の様に、コミットログの件名を `(issue_label #xxx)` で終わらせると、下の操作で CHANGELOG に掲載されるようにしている。  
     ```
    help オプションを明示的に表示させた (enhancement #2)
    ```
3. `make push-release` する。  
    以下が行われる。**途中で vi の操作を要求される。全自動ではない。**
	* CHANGELOG を更新してコミットする。  
      `_tool/add-changelog.sh` スクリプトを実行すると、以下が行われる。  
      - 前回リリース以降からスクリプト実行時までの上述のコミットログを使って、CHANGELOG を `vi` で開いてくれる。  
        体裁を整えて保存する。
      - CHANGELOG に変更があれば CHANGELOG がコミットされる。  
        ここがリリースポイント。  
        ついでに、上述のコミットログに記載されたイシュー番号をクローズするようにしている。  
        コミットログは `vi` で開いてくれるので編集可能。
    * master ブランチにマージしてプッシュする。
    * リリースタグを切ってプッシュする。  
      `_tool/add-release-tag.sh x.y.z` と引数にリリースバージョンを指定してスクリプトを実行すると、CHANGELOG から指定したバージョンの変更履歴を抜き出して注釈付きタグを作成、リモートにプッシュしてくれる。
4. `make release` する。  
   リモートにある最新のタグを使ってバイナリがリリースされる。
