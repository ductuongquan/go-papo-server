// Copyright (c) 2015-present Ladifire, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
// +build !linux

package upgrader

func CanIUpgradeToE0() error {
	return &InvalidArch{}
}

func UpgradeToE0() error {
	return &InvalidArch{}
}

func UpgradeToE0Status() (int64, error) {
	return 0, &InvalidArch{}
}
