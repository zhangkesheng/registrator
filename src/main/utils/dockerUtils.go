package utils

import (
	"os"
	"strings"
)

func CheckContainerIgnore(from string) bool {
	envIgnoreContainers := os.Getenv("IGNORE_CONTAINER")
	if strings.Contains(envIgnoreContainers, ",") {
		containers := strings.Split(envIgnoreContainers, ",")
		for _, v := range containers {
			if strings.Contains(from, v) {
				return true
			}
		}
	}
	return false
}
