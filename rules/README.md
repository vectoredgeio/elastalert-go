# Rules in ElastAlert-Go

ElastAlert-Go is a system designed for alerting and visualizing data from OpenSearch. It offers various types of rules to detect specific conditions in your data and trigger alerts accordingly.

## Rule Definitions

### 1. Any Rule

- **Type**: any
- **Description**: Triggers an alert for any event that matches the specified conditions. This rule is versatile and can be used to detect any type of event based on configurable filters.

### 2. Blacklist Rule

- **Type**: blacklist
- **Description**: Matches events based on a blacklist of values in a specific field (`compare_key`). If the field value matches any item in the blacklist, it triggers an alert. You can define blacklist values directly in the rule configuration or load them from external flat files.

### 3. Whitelist Rule

- **Type**: whitelist
- **Description**: Matches events based on a whitelist of values in a specific field (`compare_key`). Only events with values present in the whitelist trigger an alert. Useful for monitoring specific allowed activities or entities.

### 4. Change Rule

- **Type**: change
- **Description**: Triggers an alert if there is a change in a specified field (`compare_key`) between current and previous events. It compares the field value in the current event with that in the previous event to detect changes, useful for tracking modifications over time.

### 5. Frequency Rule

- **Type**: frequency
- **Description**: Triggers an alert if the frequency of events exceeds a specified threshold (`num_events`) within a defined timeframe (`timeframe`). Useful for monitoring high-volume events that exceed expected rates over specific periods.

### 6. Spike Rule

- **Type**: spike
- **Description**: Triggers an alert if there is a spike in the number of events (`spike_height`) compared to a baseline within a window (`spike_window`). It detects sudden increases in event counts, indicating potential anomalies or significant changes in activity.

### 7. Flatline Rule

- **Type**: flatline
- **Description**: Triggers an alert if no events are received within a specified timeframe (`threshold_time`). Useful for detecting periods of inactivity or when expected events cease, indicating potential service disruptions or issues.

### 8. New Term Rule

- **Type**: new_term
- **Description**: Matches events based on the occurrence of a new term in a specified field (`new_term_field`). It triggers an alert when a new term appears in the data, useful for detecting emerging trends or anomalies not previously observed.

### 9. Metric Aggregation Rule

- **Type**: metric_aggregation
- **Description**: Triggers an alert based on aggregated metrics (e.g., average, sum) exceeding a specified threshold (`aggregation_threshold`) in a specified field (`aggregation_field`). It allows monitoring aggregated data trends, such as average response times or total sales volume, to detect performance deviations or significant changes.

### 10. Spike Aggregation Rule

- **Type**: spike_aggregation
- **Description**: Triggers an alert based on aggregated spike detection in event metrics (`spike_height`) compared to a previous period. It detects sudden increases in aggregated metrics, indicating potential anomalies or significant changes in data patterns over time.

### 12. Percentage Match Rule

- **Type**: percentage_match
- **Description**: Triggers an alert if the percentage of matching events exceeds a specified threshold (`percentage_threshold`). It helps in monitoring the prevalence or distribution of specific events or conditions within the data, useful for identifying critical trends or anomalies based on percentage criteria.

Each rule type allows you to customize alerting behavior by configuring specific parameters in YAML files. These configurations define how ElastAlert-Go monitors and alerts on your data based on the rules defined.

For detailed usage examples and configuration details, please refer to the ElastAlert-Go documentation or specific rule YAML files provided with your installation.
