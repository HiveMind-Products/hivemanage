import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";
import { cn } from "@/lib/utils";
import {
  FilterSelection,
  Field,
  OPERATOR_TYPE,
  OperatorType,
} from "@/typings/logs";
import { ChevronRight, Plus } from "lucide-react";
import { useState } from "react";
import { BaseAutosuggest, isOperatorOnly } from "./base-autosuggest";
import {
  MenuOption,
  MenuOptionState,
  MenuOptionTitle,
  MultiMenu,
  OptionList,
} from "./multi-menu";

interface QueryBuilderDropdownProps {
  hasFilter?: boolean;
  metadataKeys: Field[] | null | undefined;
  onFilterSelect: (selection: FilterSelection, groupClause?: string) => void;
  isRoot?: boolean;
}

type MenuOptionType = Field | OperatorType;

const GROUP_CLAUSE = ["and", "or"];

function formatGroupClause(groupClause: string | null) {
  return `${groupClause?.toUpperCase()}(...)`;
}

export function QueryBuilderDropdown({
  metadataKeys,
  hasFilter,
  onFilterSelect,
  isRoot,
}: QueryBuilderDropdownProps) {
  // fuck these states
  const [field, setField] = useState<Field | null>(null);
  const [operator, setOperator] = useState<string | undefined>(undefined);
  const [groupClause, setGroupClause] = useState<string | undefined>(undefined);
  const [visible, setVisible] = useState(false);

  const onVisibleChange = (visible: boolean) => {
    setVisible(visible);
  };

  const onClear = () => {
    setField(null);
    setOperator(undefined);
    setGroupClause(undefined);
  };

  const handleOpenChange = (open: boolean) => {
    if (!visible) {
      onClear();
    }

    onVisibleChange(open);
  };

  const filterSelect = (selection: FilterSelection, groupClause?: string) => {
    onVisibleChange(false);
    onFilterSelect(selection, groupClause);
  };

  function renderFieldMenu({
    leftSelection,
    getOptionsProps,
  }: MenuOptionState<Field>) {
    if (metadataKeys) {
      return (
        <OptionList>
          <MenuOptionTitle>
            <p className="text-xs px-1.5">SELECT FIELD</p>
          </MenuOptionTitle>
          {metadataKeys.map((metadata) => (
            <MenuOption
              key={metadata.field}
              {...getOptionsProps(metadata)}
              className={cn(
                "p-1.5 hover:bg-accent flex justify-between items-center",
                metadata.field === leftSelection?.field && "bg-accent",
              )}
            >
              <p className="text-xs">{metadata?.field} ... </p>
              {metadata?.field === leftSelection?.field && (
                <ChevronRight size={18} className="text-muted-foreground" />
              )}
            </MenuOption>
          ))}
        </OptionList>
      );
    }
  }

  function renderOperatorMenu({
    getOptionsProps,
  }: MenuOptionState<OperatorType>) {
    return OPERATOR_TYPE.map((op) => {
      return (
        <div
          key={op}
          className="flex items-center hover:bg-accent p-1.5 pl-2"
          {...getOptionsProps(op)}
        >
          <p className="text-xs">{op}</p>
        </div>
      );
    });
  }

  function onClauseSelect(clause: string) {
    if (groupClause === clause) {
      return setGroupClause(undefined);
    }

    setGroupClause(clause);
  }

  function onMenuSelect(leftSelection: MenuOptionType, rightSelection: string) {
    if (rightSelection && leftSelection) {
      const field = leftSelection as Field;
      const op = rightSelection as OperatorType;

      if (op && isOperatorOnly(op)) {
        onFilterSelect(
          {
            field: field.field,
            operator: op,
            value: "",
            type: field.type,
          },
          groupClause!,
        );

        setGroupClause(undefined);
        onVisibleChange(false);
        return;
      }

      setField(field);
      setOperator(op);
    }
  }

  return (
    <DropdownMenu open={visible} onOpenChange={handleOpenChange} modal={false}>
      <DropdownMenuTrigger asChild className="">
        <button
          className={cn(
            "text-muted-foreground text-sm",
            !hasFilter
              ? "text-left w-full px-2 cursor-pointer"
              : "rounded-md p-0.5 cursor-pointer ",
            isRoot && hasFilter && "border hover:bg-accent",
          )}
        >
          {hasFilter ? (
            <Plus
              size={18}
              className={cn(
                isRoot
                  ? "text-foreground"
                  : "text-success",
              )}
            />
          ) : (
            "Click to add filter..."
          )}
        </button>
      </DropdownMenuTrigger>
      <DropdownMenuContent
        align="start"
        className="rounded-lg p-0 w-[600px] overflow-hidden"
      >
        <div className="mt-2 relative">
          <BaseAutosuggest
            fields={metadataKeys}
            onFilterSelect={filterSelect}
            field={field?.field}
            operator={operator}
            groupClause={groupClause}
          />

          {isRoot && (
            <div className="p-1 flex space-x-2 border-b">
              <ToggleGroup type="single" onValueChange={onClauseSelect}>
                {GROUP_CLAUSE.map((clause) => (
                  <ToggleGroupItem
                    key={clause}
                    size="sm"
                    variant="outline"
                    value={clause}
                  >
                    {formatGroupClause(clause)}
                  </ToggleGroupItem>
                ))}
              </ToggleGroup>
            </div>
          )}

          {field && operator ? null : (
            <MultiMenu<MenuOptionType>
              renderLeftMenu={renderFieldMenu}
              renderRightMenu={renderOperatorMenu}
              onSelect={onMenuSelect}
            />
          )}
        </div>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
