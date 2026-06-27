import { cn } from "@/lib/utils";
import {
  type FilterSelection,
  type FilterValue,
  type Field,
  OPERATOR_TYPE,
  OperatorType,
} from "@/typings/logs";
import flatten from "lodash/flatten";
import sortBy from "lodash/sortBy";
import { useEffect, useMemo, useState } from "react";
import {
  RenderSuggestionsContainerParams,
  SuggestionSelectedEventData,
} from "react-autosuggest";
import { Autocomplete } from "./autocomplete";
import { FilterScanner } from "./filter-scanner";
import { formatFilterValue } from "./format";
import { isNumeric, isValidOpAndFieldType, searchSort } from "./util";

interface FilterSuggestion {
  field: string;
  type: string;
  operator?: OperatorType;
  value?: FilterValue;
  rawValue?: string;
}

interface BaseAutosuggestProps {
  fields: Field[] | null | undefined;
  onFilterSelect: (selection: FilterSelection, groupClause?: string) => void;
  field?: string;
  operator?: OperatorType | undefined;
  value?: FilterValue;
  groupClause?: string;
  isEditing?: boolean;
  onCancel?: () => void;
}

export function isOperatorOnly(operator: OperatorType) {
  return operator === "exists" || operator === "not-exists";
}

const formatFieldSuggestion = (field: string) => {
  return field.indexOf(" ") === -1 ? field : `"${field}"`;
};

export function BaseAutosuggest(props: BaseAutosuggestProps) {
  const {
    field: propField,
    operator: propOperator,
    value: propValue,
    fields,
    onFilterSelect,
  } = props;

  const { _fieldStr, _fieldType } = useMemo(() => {
    const fieldStr = fields?.find((f) => f.field === propField);

    const _fieldStr = fieldStr?.field || "";
    const _fieldType = fieldStr?.type || "String";
    return { _fieldStr, _fieldType };
  }, [propField, fields]);

  const getSuggestionValue = (suggestion: FilterSuggestion) => {
    const isOperator =
      suggestion.operator && isOperatorOnly(suggestion.operator);
    const sParts = [
      formatFieldSuggestion(suggestion.field),
      suggestion.operator,
      isOperator ? undefined : formatFilterValue(suggestion.value),
    ].filter((elem) => elem !== undefined);
    let sValue = sParts.join(" ");
    if (!suggestion.operator || (sParts.length < 3 && !isOperator)) {
      sValue = `${sValue} `; // Add a trailing space if the filter is incomplete
    }

    return sValue;
  };

  const [defaultValue, setDefaultValue] = useState("");

  // if the user selects a field to edit, we need to prepopulate the operator
  useEffect(() => {
    if (propField) {
      const s = getSuggestionValue({
        field: _fieldStr,
        operator: propOperator,
        type: _fieldType,
        value: propValue,
      });

      setDefaultValue(s);
    }
  }, [propField, propOperator, _fieldStr, _fieldType, propValue]);

  const [suggestions, setSuggestions] = useState<FilterSuggestion[]>([]);

  const renderSuggestion = (
    suggestion: FilterSuggestion,
    { isHighlighted }: { isHighlighted: boolean },
  ) => {
    const parts = [suggestion.field];
    parts.push(suggestion.operator ? suggestion.operator.toString() : "...");
    if (suggestion.operator && !isOperatorOnly(suggestion.operator)) {
      parts.push(formatFilterValue(suggestion.value) || "...");
    }

    return (
      <div
        className={cn("w-full text-xs hover:bg-accent", {
          "bg-accent": isHighlighted,
        })}
      >
        <p className="p-1.5">{parts.join(" ")}</p>
      </div>
    );
  };

  const sortSuggestions = (
    { field, op }: { field?: string; op?: string },
    suggestionsToSort: FilterSuggestion[],
  ) => {
    if (!field) {
      return sortBy(suggestionsToSort, [
        (s) => s.field,
        (s) => s.operator!.toString(),
      ]);
    }
    const sorted = [...suggestionsToSort];
    const sortByField = searchSort<string>(field);
    const sortByOp = searchSort<FilterSuggestion>(op || "", "op");
    sorted.sort((a, b) => {
      if (a.field !== b.field) {
        return sortByField(a.field, b.field);
      }

      return sortByOp(a, b);
    });

    return sorted;
  };

  const onSuggestionsFetchRequested = (params: { value: string }) => {
    const trimmedValue = params.value?.trim?.() || "";
    const scanner = new FilterScanner(trimmedValue);
    scanner.scan();

    const { field, operator, value, rawValue } = scanner;

    let suggestionList: FilterSuggestion[] = [];
    if (!fields) {
      setSuggestions([]);
      return;
    }

    suggestionList = [
      ...fields.map((field) => ({
        field: field.field,
        type: field.type,
        fieldType: field.type,
      })),
    ];

    suggestionList = flatten(
      suggestionList.map((s) =>
        flatten(
          OPERATOR_TYPE.filter((o) => {
            return isValidOpAndFieldType(o, s.type);
          }).map((o) => {
            const values: FilterSuggestion[] = [];
            if (rawValue === undefined) {
              values.push({
                ...s,
                operator: o,
                value: undefined,
              });
            }
            if (value !== undefined) {
              values.push({
                ...s,
                operator: o,
                value: value,
              });
            }
            if (s.type.includes("String")) {
              if (
                rawValue !== value &&
                rawValue !== undefined &&
                rawValue.match(/["']/g) === null
              ) {
                values.push({
                  ...s,
                  operator: o,
                  value: rawValue,
                });
              }
            }

            return values;
          }),
        ),
      ),
    );

    const lowerCaseField = field?.toLocaleLowerCase();
    suggestionList = suggestionList
      .filter((s) => {
        return (
          !lowerCaseField ||
          s.field.toLocaleLowerCase().indexOf(lowerCaseField) !== -1
        );
      })
      .filter((s) => {
        if (s.value === undefined && rawValue === undefined) {
          return true;
        }
        if (
          s.operator &&
          !isOperatorOnly(s.operator) &&
          s.value !== undefined
        ) {
          let keep = false;
          if (
            s.value === null &&
            (s.operator === "==" || s.operator === "!=")
          ) {
            keep = true;
          }
          if (!keep) {
            keep =
              s.type === "any" ||
              s.type === "dynamic" ||
              s.type === "array" ||
              (s.type.includes("float") && isNumeric(s.value)) ||
              (s.type.includes("integer") && Number.isInteger(s.value)) ||
              (s.type.includes("Numeric") && Number.isInteger(s.value)) ||
              (s.type.includes("String") && typeof s.value === "string") ||
              (s.type.includes("boolean") && typeof s.value === "boolean");
          }

          return keep;
        }

        return false;
      })
      .filter((s) => {
        return operator === undefined || `${s.operator}`.startsWith(operator);
      });

    suggestionList = sortSuggestions(
      { field: field, op: operator },
      suggestionList,
    );

    setSuggestions(suggestionList);
  };

  // Autosuggest will call this function every time you need to clear suggestions.
  const onSuggestionsClearRequested = () => {
    setSuggestions([]);
  };

  const onSuggestionSelected = (
    data: SuggestionSelectedEventData<FilterSuggestion>,
  ) => {
    const { suggestion } = data;
    if (suggestion.value !== undefined && suggestion.operator) {
      onFilterSelect(
        {
          field: suggestion.field,
          operator: suggestion.operator,
          value: suggestion.value,
          type: suggestion.type,
        },
        props.groupClause,
      );
    } else if (suggestion.operator && isOperatorOnly(suggestion.operator)) {
      onFilterSelect(
        {
          field: suggestion.field,
          operator: suggestion.operator,
          value: null,
          type: suggestion.type,
        },
        props.groupClause,
      );
    }
  };

  return (
    <Autocomplete
      defaultValue={defaultValue}
      suggestions={suggestions}
      onSuggestionsFetchRequested={onSuggestionsFetchRequested}
      onSuggestionsClearRequested={onSuggestionsClearRequested}
      getSuggestionValue={getSuggestionValue}
      renderSuggestion={renderSuggestion}
      renderSuggestionsContainer={renderSuggestionsContainer}
      onSuggestionSelected={onSuggestionSelected}
      onCancel={props.onCancel}
    />
  );
}

function renderSuggestionsContainer({
  containerProps,
  children,
}: RenderSuggestionsContainerParams) {
  return (
    <div
      {...containerProps}
      className="w-full mt-2 p-0 max-h-96 overflow-y-auto border  rounded-md bg-background shadow-lg"
    >
      {children}
    </div>
  );
}
