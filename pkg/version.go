package pkg

var (
	//GitCommit Git Commit SHA
	GitCommit string
	//Version version of the CLI
	Version string
)

//GetVersion get latest version
func GetVersion() string {
	if len(Version) == 0 {
		return "dev"
	}
	return Version
}

const Logo = `  __                     _ 
 / _| __ _  __ _ ___  __| |
| |_ / _` + "`" + ` |/ _` + "`" + ` / __|/ _` + "`" + ` |
|  _| (_| | (_| \__ \ (_| |
|_|  \__,_|\__,_|___/\__,_|
`
