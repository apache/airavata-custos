package operations

import "os"

func isLocalSlurmConfigAvailable() bool {
	if os.Getenv("TEST_SLURM_API") == "" || os.Getenv("TEST_SLURM_USER") == "" || os.Getenv("TEST_SLURM_TOKEN") == "" || os.Getenv("TEST_SLURM_API_VERSION") == "" {
		return false
	}
	return true
}
