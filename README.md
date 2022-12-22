# Weather
#### A program for keeping track of weather.

# Install

To install:

```shell
$ go install github.com/unclassedpenguin/weather@latest
```

Then, you need to create the config file at `~/.config/weather/config.yaml`

### Example ~/.config/weather/config.yaml

```yaml
# Database dir is the directory you want to store your databases in.
# It can be a git repo, but doesn't have to be...
# This directory must be created by the user
DatabaseDir: /home/username/git/databases

# RealDatabase is the legit database
# This will be created if it doesn't exist
RealDatabase: weatherDatabase.db

# TestDatabase is a database you can use to test features
# This will be created if it doesn't exist
TestDatabase: weatherTestDatabase.db

```

Don't forget to edit the DatabaseDir to where you want to store the databases. You just have to create the folder, the databases will be created if they don't exist.


## To-do:
 
 - Add ability to sort asc or desc based on high or low temp. 

 #### v0.0.1
