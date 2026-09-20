package buildinfo

import "runtime/debug"

const developmentVersion = "dev"

func Version(linkedVersion string) string {
	moduleVersion := ""
	if info, ok := debug.ReadBuildInfo(); ok {
		moduleVersion = info.Main.Version
	}
	return resolveVersion(linkedVersion, moduleVersion)
}

func resolveVersion(linkedVersion, moduleVersion string) string {
	if linkedVersion != "" && linkedVersion != developmentVersion {
		return linkedVersion
	}
	if moduleVersion != "" && moduleVersion != "(devel)" {
		return moduleVersion
	}
	return developmentVersion
}
