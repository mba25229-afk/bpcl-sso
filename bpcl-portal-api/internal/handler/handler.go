package handler

type Handler struct {
	Auth         AuthServiceI
	Outlet       OutletServiceI
	Performance  PerformanceServiceI
	Target       TargetServiceI
	Upload       UploadServiceI
	Competition  CompetitionServiceI
	MarketShare  MarketShareServiceI
	Users       UserServiceI
}

func New(
	auth AuthServiceI,
	outlet OutletServiceI,
	perf PerformanceServiceI,
	target TargetServiceI,
	upload UploadServiceI,
	competition CompetitionServiceI,
	marketShare MarketShareServiceI,
	users UserServiceI,
) *Handler {
	return &Handler{
		Auth:        auth,
		Outlet:      outlet,
		Performance: perf,
		Target:      target,
		Upload:      upload,
		Competition: competition,
		MarketShare: marketShare,
		Users:       users,
	}
}
