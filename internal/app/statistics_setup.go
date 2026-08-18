package app

import (
	"github.com/txyyddss/Remna-User-Panel/internal/providerops"
	productstats "github.com/txyyddss/Remna-User-Panel/internal/statistics"
)

func registerStatisticsOperationHandler(dispatcher *providerops.Dispatcher, repository productstats.HostWorkerRepository, provider productstats.Provider) error {
	worker := productstats.NewHostRemarkWorker(repository, provider)
	return dispatcher.Register(providerops.KindHostRemarkUpdate, worker)
}
