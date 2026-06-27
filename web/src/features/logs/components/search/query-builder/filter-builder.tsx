import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Label } from "@/components/ui/label";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";
import { Filter, FilterSelection, Field } from "@/typings/logs";
import { ChevronDown, X } from "lucide-react";
import { useState } from "react";
import { BaseAutosuggest, isOperatorOnly } from "./base-autosuggest";
import { QueryBuilderDropdown } from "./query-builder-dropdown";
import { formatFilterValue } from "./format";

interface FilterBuilderProps {
  filter: Filter;
  fields: Field[] | null | undefined;
  onFilterChange: (filter: Filter) => void;
  hideClauseDropdown?: boolean;
}

export function selectionToFilter(
  selection: FilterSelection,
  groupClause?: string,
): Filter {
  const { field, operator, value, type } = selection;

  if (groupClause) {
    const child: Filter = {
      operator: operator,
      value: value,
      type: type,
      field: field,
    };

    return {
      operator: groupClause,
      field: "",
      children: [child as Filter],
    };
  }

  const filter: Filter = {
    field: field,
    operator: operator,
    value: value,
    type: type,
  };

  return filter;
}

export function FilterBuilder(props: FilterBuilderProps) {
  const [editingFilter, setEditingFilter] = useState<Filter | undefined>(
    undefined,
  );
  const { filter, fields, hideClauseDropdown, onFilterChange } = props;

  function onFilterSelect(
    filter_: Filter,
    selection: FilterSelection,
    groupClause?: string,
  ) {
    const selectedFilter = selectionToFilter(selection, groupClause);

    if (selectedFilter) {
      let updatedFilter = props.filter;
      if (updatedFilter) {
        updatedFilter = add(updatedFilter, filter_, selectedFilter);
      }

      onFilterChange(updatedFilter);
    }
  }

  function onEditFilter(filter_: Filter) {
    setEditingFilter(filter_);
  }

  function clearEditing() {
    setEditingFilter(undefined);
  }

  function onRemoveFilter(filter_: Filter) {
    const updatedFilter = remove(filter, filter_);
    onFilterChange(updatedFilter);
  }

  function onFilterOpChange(filter_: Filter, operator: string) {
    // todo: mayyybe we need to add the operator back
    function doSet(f: Filter): Filter {
      if (f === filter_) {
        return { ...f, operator: operator };
      }

      if (f.children) {
        f.children = f.children.map((child) => doSet(child));

        return f;
      }

      return f;
    }

    const newFilter = doSet(filter);
    onFilterChange(newFilter);
  }

  function updateFilter(filter: Filter, child: Filter) {
    function doUpdate(f: Filter) {
      if (f === filter) {
        return { ...child, children: f.children };
      }

      if (f.children) {
        f.children = f.children.map(doUpdate);

        return f;
      }

      return f;
    }

    const currFilter = props.filter;
    const newFilter = doUpdate(currFilter);

    setEditingFilter(undefined);
    props.onFilterChange(newFilter);
  }

  function add(root: Filter, target: Filter, child: Filter) {
    function addChild(f: Filter) {
      if (f === target) {
        const newChildren = f.children ? [...f.children, child] : [child];

        f.children = newChildren;

        return;
      }

      if (f.children) {
        f.children.forEach(addChild);

        return;
      }
    }

    addChild(root);

    return { ...root };
  }

  function remove(root: Filter, target: Filter): Filter {
    function removeChild(node: Filter) {
      if (node === target) {
        return null;
      }

      if (node.children) {
        node.children = node.children
          .map(removeChild)
          .filter((child) => child !== null);
      }
      return node;
    }

    const newRoot = removeChild(root);
    if (newRoot === null) {
      return { ...root, children: [] };
    }

    return newRoot;
  }

  return (
    <div className="flex items-center space-x-2">
      <InternalFilterBuilder
        filter={filter}
        fields={fields}
        hideClauseDropdown={hideClauseDropdown}
        removeFilter={onRemoveFilter}
        setFilterOp={onFilterOpChange}
        onFilterSelect={onFilterSelect}
        editFilter={onEditFilter}
        editingFilter={editingFilter}
        clearEditing={clearEditing}
        updateFilter={updateFilter}
        isRoot
      />
    </div>
  );
}

interface InternalFilterBuilderProps {
  filter: Filter;
  fields: Field[] | null | undefined;
  removeFilter: (filter: Filter) => void;
  hideClauseDropdown?: boolean;
  setFilterOp(filter: Filter, op: string): void;
  onFilterSelect: (
    filter: Filter,
    selection: FilterSelection,
    groupClause?: string,
  ) => void;
  dropdownKey?: string;
  isRoot?: boolean;
  editingFilter?: Filter;
  editFilter(filter: Filter): void;
  clearEditing(): void;
  updateFilter(filter: Filter, child: Filter): void;
}

function InternalFilterBuilder(props: InternalFilterBuilderProps) {
  const { filter, removeFilter, fields } = props;

  const { operator } = filter;

  if (operator === "and" || operator === "or") {
    return renderGroupFilter();
  }

  const { field } = filter;

  if (field) {
    return renderFilter();
  }

  return null;

  function onFilterOpChange(value: string) {
    const { filter } = props;
    props.setFilterOp(filter, value);
  }

  function onFilterSelect(
    filter: Filter,
    selection: FilterSelection,
    groupClause?: string,
  ) {
    props.onFilterSelect(filter, selection, groupClause);
  }

  function onRemoveFilter(filter: Filter) {
    removeFilter(filter);
  }

  function onUpdateFilter(newFilter: Filter) {
    const { field, operator, value } = newFilter;

    if (isOperatorOnly(operator)) {
      props.updateFilter(props.filter, {
        field: field,
        operator: operator,
        type: "metadata",
      });
    } else if (value !== undefined) {
      props.updateFilter(props.filter, {
        field: field,
        operator: operator,
        value,
        type: "metadata",
      });
    }
  }

  function onAutosuggestCancel() {
    props.clearEditing();
  }

  function renderGroupFilter() {
    const { children, operator } = filter;

    const mappedChildren = children
      ? children.map((queryFilter, idx) => {
          const id = `${idx}-${queryFilter.field ? queryFilter.field.replace(" ", "") : ""}`;

          return (
            <InternalFilterBuilder
              key={id}
              {...props}
              dropdownKey={id}
              filter={queryFilter}
              fields={props.fields}
              isRoot={false}
              removeFilter={(filter) => onRemoveFilter(filter)}
              hideClauseDropdown={false}
              setFilterOp={props.setFilterOp}
              onFilterSelect={(_, selection, groupClause) => {
                onFilterSelect(queryFilter, selection, groupClause);
              }}
            />
          );
        })
      : null;

    return (
      <div
        className={cn(
          "flex items-center space-x-1 rounded-sm px-2 text-sm shrink-0",
          props.hideClauseDropdown
            ? "bg-transparent"
            : "bg-green-50 dark:bg-green-900",
        )}
      >
        {props.hideClauseDropdown ? null : (
          <GroupClauseDropdown
            key={`groupClause-${props.dropdownKey}`}
            operator={operator}
            onRemove={() => onRemoveFilter(filter)}
            onChange={onFilterOpChange}
          />
        )}
        {mappedChildren}
        <QueryBuilderDropdown
          hasFilter
          isRoot={props.isRoot}
          metadataKeys={fields}
          onFilterSelect={(selection, groupClause) => {
            // since we this is a recursive function, we need to pass the groupCla
            onFilterSelect(props.filter, selection, groupClause);
          }}
        />
      </div>
    );
  }

  function renderFilter() {
    const { operator, field, value } = filter;

    const isEditing = !!props.editingFilter;

    if (isEditing && filter === props.editingFilter) {
      return (
        <div className="absolute w-full z-999 inset-0 bg-background">
          <BaseAutosuggest
            isEditing={isEditing}
            field={field}
            operator={operator}
            value={value}
            fields={fields}
            onFilterSelect={(selection) => {
              onUpdateFilter(selection);
            }}
            onCancel={onAutosuggestCancel}
          />
        </div>
      );
    }

    return isEditing ? null : (
      <div
        className={cn(
          "flex rounded-md h-7 text-sm text-green-800 dark:text-green-100 shrink-0 bg-green-50 dark:bg-green-900",
        )}
      >
        <div className="flex items-center justify-center px-1 cursor-pointer">
          <TooltipProvider>
            <Tooltip delayDuration={0}>
              <TooltipTrigger className="inline-flex" asChild>
                <button
                  className="cursor-pointer"
                  onClick={() => {
                    onRemoveFilter(props.filter);
                  }}
                >
                  <X size={14} />
                </button>
              </TooltipTrigger>
              <TooltipContent side="bottom" className="text-xs">
                Remove {field} {operator} &quot;
                {value}
                &quot;
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        </div>
        <p
          className="whitespace-nowrap text-xs cursor-pointer flex items-center px-1"
          onClick={() => {
            onFilterClick(props.filter);
          }}
        >
          {field} {operator}{" "}
          {isOperatorOnly(operator) ? "" : formatFilterValue(value)}
        </p>
      </div>
    );
  }

  function onFilterClick(filter: Filter) {
    const { field, operator, value } = filter;

    if (field && operator && value) {
      props.editFilter(filter);
    }
  }
}

interface GroupClauseDropdownProps {
  operator: string;
  onRemove: () => void;
  onChange: (value: string) => void;
}

function GroupClauseDropdown({
  operator,
  onChange,
  onRemove,
}: GroupClauseDropdownProps) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="link"
          size="sm"
          className="text-green-800 dark:text-green-100 h-6"
        >
          {operator.toUpperCase()}
          <ChevronDown size={12} />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent className="p-2 px-4">
        <RadioGroup value={operator} onValueChange={onChange}>
          <div className="flex space-x-4">
            <div id="group-and" className="flex items-center gap-x-2">
              <RadioGroupItem value="and" id="group-and" />
              <Label htmlFor="group-and">AND</Label>
            </div>
            <div id="group-or" className="flex items-center gap-x-2">
              <RadioGroupItem value="or" id="group-or" />
              <Label htmlFor="group-or">OR</Label>
            </div>
          </div>
        </RadioGroup>

        <div className="mt-2">
          <Button
            variant="destructive"
            size="sm"
            className="w-full"
            onClick={onRemove}
          >
            Remove group
          </Button>
        </div>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
