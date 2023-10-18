// SPDX-License-Identifier: Apache-2.0

package secretsapp

type Store struct{}

func (s *Store) Get(name string) ([]byte, error) {
	return nil, nil
}

func (s *Store) Set(name string, value []byte) error {
	return nil
}
