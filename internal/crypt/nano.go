package crypt

import gonanoid "github.com/matoous/go-nanoid/v2"

func GenerateSessionID() (string, error) {
	id, err := gonanoid.Generate("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ_abcdefghijklmnopqrstuvwxyz", 24)
	if err != nil {
		return "", err
	}
	return id, nil
}

func GenerateApiKey() (string, error) {
	apiKey, err := gonanoid.Generate("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ_abcdefghijklmnopqrstuvwxyz", 32)
	if err != nil {
		return "", err
	}
	return apiKey, nil
}

func GenerateFilename() (string, error) {
	id, err := gonanoid.Generate("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz", 24)
	if err != nil {
		return "", err
	}
	return id, nil
}

func GeneratePrimaryKey() (string, error) {
	id, err := gonanoid.Generate("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz", 16)
	if err != nil {
		return "", err
	}
	return id, nil
}

func GeneratePassword() (string, error) {
	return gonanoid.Generate("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz!@#$%", 16)
}
