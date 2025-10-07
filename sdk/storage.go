package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/dgraph-io/badger/v4"
)

type storage interface {
	persistParameters(ctx context.Context, parameters []Parameter) error
	getParameterByName(ctx context.Context, name string) (Parameter, error)
	persistExperiments(ctx context.Context, experiments []Experiment) error
	close(ctx context.Context) error
	getExperimentsByParameterName(ctx context.Context, parameterName string) ([]Experiment, error)
}

type storageImpl struct {
	db     *badger.DB
	logger *slog.Logger
}

func newStorage(db *badger.DB, logger *slog.Logger) storage {
	return &storageImpl{
		db:     db,
		logger: logger,
	}
}

func (s *storageImpl) persistParameters(ctx context.Context, parameters []Parameter) error {
	for _, parameter := range parameters {
		jsonParameters, err := json.Marshal(parameter)
		if err != nil {
			return NewStorageError("marshal parameter", err)
		}
		err = s.db.Update(func(txn *badger.Txn) error {
			return txn.Set([]byte(fmt.Sprintf("parameters:%s", parameter.Name)), jsonParameters)
		})
		if err != nil {
			return NewStorageError("store parameter", err)
		}
	}
	return nil
}

func (s *storageImpl) getParameterByName(ctx context.Context, name string) (Parameter, error) {
	var parameter Parameter
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(fmt.Sprintf("parameters:%s", name)))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &parameter)
		})
	})
	if err != nil {
		return Parameter{}, NewStorageError("get parameter", err)
	}
	return parameter, nil
}

func (s *storageImpl) persistExperiments(ctx context.Context, experiments []Experiment) error {
	mapParameters := make(map[string][]string)

	for _, experiment := range experiments {
		jsonExperiments, err := json.Marshal(experiment)
		if err != nil {
			return NewStorageError("marshal experiment", err)
		}
		err = s.db.Update(func(txn *badger.Txn) error {
			return txn.Set([]byte(experiment.Name), jsonExperiments)
		})
		if err != nil {
			return NewStorageError("store experiment", err)
		}
	}

	for _, experiment := range experiments {
		for _, variant := range experiment.Variants {
			for _, parameter := range variant.Parameters {
				mapParameters[parameter.ParameterName] = append(mapParameters[parameter.ParameterName], experiment.Name)
			}
		}
	}

	for k, v := range mapParameters {
		jsonParameters, err := json.Marshal(v)
		if err != nil {
			return NewStorageError("marshal parameters", err)
		}
		err = s.db.Update(func(txn *badger.Txn) error {
			return txn.Set([]byte(fmt.Sprintf("experiments:parameters:%s", k)), jsonParameters)
		})
		if err != nil {
			return NewStorageError("store parameters", err)
		}
	}
	return nil
}

func (s *storageImpl) close(ctx context.Context) error {
	return s.db.Close()
}

func (s *storageImpl) getExperimentByName(ctx context.Context, name string) (Experiment, error) {
	var experiment Experiment
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(name))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &experiment)
		})
	})
	if err != nil {
		return Experiment{}, NewStorageError("get experiment by name", err)
	}
	return experiment, nil
}

func (s *storageImpl) getExperimentsByParameterName(ctx context.Context, parameterName string) ([]Experiment, error) {
	var experimentNames []string
	var experiments []Experiment
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(fmt.Sprintf("experiments:parameters:%s", parameterName)))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &experimentNames)
		})
	})
	if err != nil {
		return []Experiment{}, NewStorageError("get experiments by parameter name", err)
	}

	for _, experimentName := range experimentNames {
		experiment, err := s.getExperimentByName(ctx, experimentName)
		if err != nil {
			return []Experiment{}, NewStorageError("get experiments by parameter name", err)
		}
		experiments = append(experiments, experiment)
	}
	return experiments, nil
}
