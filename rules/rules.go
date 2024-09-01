package rules

import (
	anyrule "elastalert-go/rules/anyRule"
	blacklistrule "elastalert-go/rules/blacklistRule"
	cardinalityrule "elastalert-go/rules/cardinalityRule"
	changerule "elastalert-go/rules/changeRule"
	flatlinerule "elastalert-go/rules/flatlineRule"
	frequencyrule "elastalert-go/rules/frequencyRule"
	metricaggregationrule "elastalert-go/rules/metricAgrregationRule"
	newtermrule "elastalert-go/rules/newTermRule"
	percentagematchrule "elastalert-go/rules/percentageMatchRule"
	spikerule "elastalert-go/rules/spikeRule"
	spikeaggregationrule "elastalert-go/rules/spikeaggregationRule"
	whitelistrule "elastalert-go/rules/whitelistRule"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/opensearch-project/opensearch-go/opensearchapi"
	"gopkg.in/yaml.v2"
)

// Rule is the interface that all rules must implement.
type Rule interface {
	// Define methods required by the Rule interface
	GetName() string
	GetIndex() string
	GetType() string
	GetQuery() (*opensearchapi.SearchRequest, error) 
	Evaluate(hits []map[string]interface{}) bool
}

// RuleFactory holds mappings of rule types to their constructors.
var RuleFactory = map[string]func() Rule{
	"any":               func() Rule { return &anyrule.AnyRule{} },
	"blacklist":             func() Rule { return &blacklistrule.BlacklistRule{} },
	"cardinality":           func() Rule { return &cardinalityrule.CardinalityRule{} },
	"change":                func() Rule { return &changerule.ChangeRule{} },
	"flatline":              func() Rule { return &flatlinerule.FlatlineRule{} },
	"frequency":             func() Rule { return &frequencyrule.FrequencyRule{} },
	"metric_aggregation":     func() Rule { return &metricaggregationrule.MetricAggregationRule{}},
	"new_term":               func() Rule { return &newtermrule.NewTermRule{} },
	"percentage_match":       func() Rule { return &percentagematchrule.PercentageMatchRule{} },
	"spike_aggregation":      func() Rule { return &spikeaggregationrule.SpikeAggregationRule{} },
	"whitelist":             func() Rule { return &whitelistrule.WhitelistRule{}},
	"spike":				func() Rule {return &spikerule.SpikeRule{}},
}

type DualEvaluatable interface {
    EvaluateDual(currentHits, previousHits []map[string]interface{}) bool
}
type EvaluateAggregations interface {
    EvaluateAggregations(aggregations map[string]interface{}) bool
}

// LoadRule dynamically loads a rule based on its YAML configuration file.
func LoadRule(filename string) (Rule, error) {
    data, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, err
    }

    var ruleConfig struct {
        Type string `yaml:"type"`
    }
    err = yaml.Unmarshal(data, &ruleConfig)
    if err != nil {
        return nil, err
    }

    constructor, ok := RuleFactory[ruleConfig.Type]
    if !ok {
        return nil, fmt.Errorf("unsupported rule type: %s", ruleConfig.Type)
    }

    rule := constructor()
    if rule == nil {
        return nil, errors.New("failed to create rule instance")
    }

    err = yaml.Unmarshal(data, rule)
    if err != nil {
        return nil, err
    }

    return rule, nil
}




func EvaluateRule(rule Rule, currentHits, previousHits []map[string]interface{}, aggregations map[string]interface{}) bool {
    if dualRule, ok := rule.(DualEvaluatable); ok {
        return dualRule.EvaluateDual(currentHits, previousHits)
    } else if aggRule, ok := rule.(EvaluateAggregations); ok {
        return aggRule.EvaluateAggregations(aggregations)
    } else {
        return rule.Evaluate(currentHits)
    }
}

