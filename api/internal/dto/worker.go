package dto

import "github.com/riverqueue/river"

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

func (SyncParameterArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: "sync_parameter",
	}
}

func (SyncExperimentArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: "sync_experiment",
	}
}
