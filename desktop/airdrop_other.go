//go:build !darwin

package main

import "errors"

const airDropSupported = false

func airDrop(string) error { return errors.New("AirDrop chỉ có trên máy Mac") }
