<!-- togo-brand -->
<p align="center"><img src=".github/assets/togo-mark.svg" width="96" alt="togo" /></p>
<h1 align="center">data-databricks</h1>
<p align="center"><sub>part of the <a href="https://github.com/togo-framework">togo-framework</a></sub></p>

A togo **data** backend that queries **Databricks SQL** (Statement Execution API).

```bash
togo install togo-framework/data-databricks
togo provider:use data databricks
togo config:set DATABRICKS_HOST https://xxx.cloud.databricks.com
togo config:set DATABRICKS_WAREHOUSE_ID abc123
# DATABRICKS_TOKEN in .env (secret)
```

MIT © fadymondy
