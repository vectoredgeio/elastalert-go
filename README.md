
# ElastAlert-Go

Go port of Elastalert 2  , ElastAlert-Go is a tool for alerting on Opensearch data. It monitors Opensearch for defined conditions and sends alerts using various notification systems.

## Project Structure

- `cmd/elastalert/`: Main application entry point
- `config/`: Configuration management
- `rules/`: Rule definitions and parsing
- `queries/`: Opensearch querying logic
- `alerts/`: Alerting mechanisms
- `processor/`: Core logic for processing results and generating alerts
- `util/`: Utility functions

## Getting Started

1. Configure `config/config.yaml`.
2. Define your rules in the `rules/` directory.
3. Run the application:

    ```sh
    go run cmd/elastalert/main.go
    ```

## Credits

This project was developed under the guidance and support of the organization VectorEdge. Their expertise and resources were instrumental in the successful completion of ElastAlert-Go. For more information, visit their [website](https://vectoredge.io/).
