# gen-core
コアシステムはGolangで構築されており、APIサービス、ウェブサービスなどの開発に使用されます。

## workflow (ワークフロー)
![](doc/assets/flow.png)

## Docker

### docker provides support for starting and stopping docker containers for running tests.
Docker は、テストを実行するための Docker コンテナの開始および停止をサポートしています。

container struct
```go
// Container tracks information about the docker container started for tests.
type Container struct {
	ID   string
	Host string // IP:Port
}
```

#### Extract IP Port docker container with 'Format command and log output' of Docker using Go template
Go テンプレートを使用した Docker の「フォーマットコマンドとログ出力」で Docker コンテナの IP ポートを抽出する
[doc](https://docs.docker.com/config/formatting/)

```go
tmpl := fmt.Sprintf("[{{range $k,$v := (index .NetworkSettings.Ports \"%s/tcp\")}}{{json $v}}{{end}}]", port)

cmd := exec.Command("docker", "inspect", "-f", tmpl, id)
var out bytes.Buffer
cmd.Stdout = &out
if err := cmd.Run(); err != nil {
	// handle err
}
```

## Keystore (キーストア)
structure

![](/doc/assets/keystore.png)
