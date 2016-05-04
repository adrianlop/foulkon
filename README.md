# authorizr

authorizr is an authorization server that allows or denies the access to web resources.

## Installation / usage


## Configuration
You have to specify configuration file using flag -config-file (authorizr -config-file=/path/configfile.json). This config file is a JSON file that has four parts:

#### LoggerConfig:
    - log_type : file | default (If it isn't specified it uses stdout)
    - log_file_dir: /path/file.log (If you select log_type file you have to specify the log dir file)
    - log_level_debug: true | false (Debug log enabled)
#### DatabaseConfig:
    - db_type : postgres (Only postgres right now)
    - db_postgres_datasourcename: dsn (Datasource name for connecting to postgres)
#### AuthConnectorConfig:
    - auth_connector_type : oidc (Only OIDC protocol right now)
    - auth_connector_oidc_issuer: www.example.com (Your selected issuer for OIDC tokens)
    - auth_connector_oidc_client_ids: clientid1;clientid2 (Client IDs that you accept separated by ",")
#### AdminUserConfig:
    - auth_admin_username : admin (Admin username)
    - auth_admin_password: password (Admin password)

This is a config file example:

```json
{
    "LoggerConfig": {
        "log_type": "file",
        "log_file_dir": "/tmp/authorizr/authorizr.log",
        "log_level_debug": "true"
    },
    "DatabaseConfig": {
        "db_type": "postgres",
        "db_postgres_datasourcename": "postgres://authorizr:password@localhost:5432/authorizrdb?sslmode=disable"
    },
    "AuthConnectorConfig": {
        "auth_connector_type": "oidc",
        "auth_connector_oidc_issuer": "http://localhost:5556",
        "auth_connector_oidc_client_ids": "9jCU4aaDHjV-y59SSlGwfrmpdo4mIkGBW4E41QvI-X0=@127.0.0.1"
    },
    "AdminUserConfig": {
        "auth_admin_username": "admin",
        "auth_admin_password": "admin"
    }
}
```

## Testing


## Contribution policy

Contributions via GitHub pull requests are gladly accepted from their original author. Along with any pull requests, please state that the contribution is your original work and that you license the work to the project under the project's open source license. Whether or not you state this explicitly, by submitting any copyrighted material via pull request, email, or other means you agree to license the material under the project's open source license and warrant that you have the legal authority to do so.

Please make sure to follow these conventions:
- For each contribution there must be a ticket (GitHub issue) with a short descriptive name, e.g. "Respect seed-nodes configuration setting"
- Work should happen in a branch named "ISSUE-DESCRIPTION", e.g. "32-respect-seed-nodes"
- Before a PR can be merged, all commits must be squashed into one with its message made up from the ticket name and the ticket id, e.g. "Respect seed-nodes configuration setting (closes #32)"

#### Questions

If you have a question, preferably use the [mailing list](mailto:dev.whiterabbit@tecsisa.com) or Google Hangouts. As an alternative, prepend your issue with `[question]`.

## License

This code is open source software licensed under the [Apache 2.0 License](http://www.apache.org/licenses/LICENSE-2.0.html).

## TODO
- DBs: PostgreSQL. Cayley?
- support OIDC token
- Logging: https://github.com/Sirupsen/logrus
