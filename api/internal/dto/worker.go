package dto

type SyncParameterArgs struct {
	ParameterID int
}

func (SyncParameterArgs) Kind() string {
	return "sync_parameter"
}
