package store

import contract "onboard-kit/contracts/preparation/store"

const byteLimit = 1048576

type store struct{}

func New() contract.Store { return store{} }
