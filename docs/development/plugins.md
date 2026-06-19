# API плагинов и расширений

Python entry point API `minishop.plugins` удален вместе со старым backend.
Новые расширения пишутся на Go и используют публичный пакет `pkg/plugin`.

Минимальный контракт:

```go
package example

import (
	"context"
	"net/http"

	"remna-user-panel/pkg/plugin"
)

type Plugin struct{}

func (Plugin) Name() string { return "example" }
func (Plugin) Version() string { return "0.1.0" }
func (Plugin) Setup(context.Context, *plugin.Context) error { return nil }
func (Plugin) RegisterHTTP(*http.ServeMux) {}
func (Plugin) WorkerTasks(*plugin.Context) []plugin.WorkerTask { return nil }
```

`plugin.Context` содержит настройки и реестр сервисов процесса. Worker tasks
получают `context.Context` и обязаны завершаться при отмене контекста.

Встроенные расширения, которые раньше жили как Python plugins, должны быть
перенесены в Go packages и подключены на этапе wiring в `internal/app`.
Совместимость с внешними Python-плагинами не сохраняется.
