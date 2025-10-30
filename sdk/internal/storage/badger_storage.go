package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"sdk/pkg/errors"
	"sdk/pkg/logger"
	"sdk/types"

	"github.com/dgraph-io/badger/v4"
)

// BadgerStorage implements Storage using BadgerDB
type BadgerStorage struct {
	db     *badger.DB
	logger logger.Logger
}

// NewBadgerStorage creates a new BadgerDB storage instance
func NewBadgerStorage(db *badger.DB, logger logger.Logger) Storage {
	return &BadgerStorage{
		db:     db,
		logger: logger,
	}
}

// PersistParameters stores parameters in the database
func (s *BadgerStorage) PersistParameters(ctx context.Context, parameters []types.Parameter) error {
	for _, parameter := range parameters {
		jsonParameters, err := json.Marshal(parameter)
		if err != nil {
			return errors.NewStorageError("marshal parameter", err)
		}
		err = s.db.Update(func(txn *badger.Txn) error {
			return txn.Set([]byte(fmt.Sprintf("parameters:%s", parameter.Name)), jsonParameters)
		})
		if err != nil {
			return errors.NewStorageError("store parameter", err)
		}
	}
	return nil
}

// GetParameterByName retrieves a parameter by name
func (s *BadgerStorage) GetParameterByName(ctx context.Context, name string) (types.Parameter, error) {
	var parameter types.Parameter
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
		return types.Parameter{}, errors.NewStorageError("get parameter", err)
	}
	return parameter, nil
}

// PersistExperiments stores experiments in the database
func (s *BadgerStorage) PersistExperiments(ctx context.Context, experiments []types.Experiment) error {
	mapParameters := make(map[string][]string)

	// Store experiments
	for _, experiment := range experiments {
		jsonExperiments, err := json.Marshal(experiment)
		if err != nil {
			return errors.NewStorageError("marshal experiment", err)
		}
		err = s.db.Update(func(txn *badger.Txn) error {
			return txn.Set([]byte(experiment.Name), jsonExperiments)
		})
		if err != nil {
			return errors.NewStorageError("store experiment", err)
		}
	}

	// Build parameter to experiment mapping
	for _, experiment := range experiments {
		for _, variant := range experiment.Variants {
			for _, parameter := range variant.Parameters {
				mapParameters[parameter.ParameterName] = append(mapParameters[parameter.ParameterName], experiment.Name)
			}
		}
	}

	// Store parameter mappings
	for k, v := range mapParameters {
		jsonParameters, err := json.Marshal(v)
		if err != nil {
			return errors.NewStorageError("marshal parameters", err)
		}
		err = s.db.Update(func(txn *badger.Txn) error {
			return txn.Set([]byte(fmt.Sprintf("experiments:parameters:%s", k)), jsonParameters)
		})
		if err != nil {
			return errors.NewStorageError("store parameters", err)
		}
	}
	return nil
}

// GetExperimentsByParameterName retrieves experiments that contain a specific parameter
func (s *BadgerStorage) GetExperimentsByParameterName(ctx context.Context, parameterName string) ([]types.Experiment, error) {
	var experimentNames []string
	var experiments []types.Experiment
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
		return []types.Experiment{}, errors.NewStorageError("get experiments by parameter name", err)
	}

	for _, experimentName := range experimentNames {
		experiment, err := s.getExperimentByName(ctx, experimentName)
		if err != nil {
			return []types.Experiment{}, errors.NewStorageError("get experiments by parameter name", err)
		}
		experiments = append(experiments, experiment)
	}
	return experiments, nil
}

// getExperimentByName retrieves an experiment by name
func (s *BadgerStorage) getExperimentByName(ctx context.Context, name string) (types.Experiment, error) {
	var experiment types.Experiment
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
		return types.Experiment{}, errors.NewStorageError("get experiment by name", err)
	}
	return experiment, nil
}

// Close closes the storage
func (s *BadgerStorage) Close(ctx context.Context) error {
	return s.db.Close()
}
