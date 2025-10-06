package dto

type SyncParameterArgs struct {
	ParameterID int
}

type SyncExperimentArgs struct {
}

func (SyncExperimentArgs) Kind() string {
	return "sync_experiment"
}

func (SyncParameterArgs) Kind() string {
	return "sync_parameter"
}
