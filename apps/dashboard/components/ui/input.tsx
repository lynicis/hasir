import * as React from "react";

import { cn } from "@/lib/utils";

type InputProps = React.ComponentProps<"input">;

type SelectionRange = {
  start: number;
  end: number;
};

const maskValue = (value: string) => (value ? "*".repeat(value.length) : "");
const resolveInitialValue = (
  value: InputProps["value"],
  defaultValue: InputProps["defaultValue"]
) =>
  typeof value === "string"
    ? value
    : typeof defaultValue === "string"
    ? defaultValue
    : "";

function useMaskedPassword({
  value,
  defaultValue,
  onChange,
  onInput,
}: Pick<InputProps, "value" | "defaultValue" | "onChange" | "onInput">) {
  const inputRef = React.useRef<HTMLInputElement>(null);
  const initialValue = resolveInitialValue(value, defaultValue);
  const actualValueRef = React.useRef(initialValue);
  const [maskedValue, setMaskedValue] = React.useState(() =>
    maskValue(initialValue)
  );

  const selectionRef = React.useRef<SelectionRange>({
    start: initialValue.length,
    end: initialValue.length,
  });

  React.useEffect(() => {
    if (typeof value === "string" && value !== actualValueRef.current) {
      actualValueRef.current = value;
      setMaskedValue(maskValue(value));
      selectionRef.current = {
        start: value.length,
        end: value.length,
      };
    }
  }, [value]);

  const emitChange = React.useCallback(
    (
      nextValue: string,
      event:
        | React.ChangeEvent<HTMLInputElement>
        | React.FormEvent<HTMLInputElement>
    ) => {
      const target = event.target as HTMLInputElement;
      const currentTarget = event.currentTarget as HTMLInputElement;

      target.value = nextValue;
      currentTarget.value = nextValue;

      if ("nativeEvent" in event && "type" in event) {
        if (event.type === "change") {
          onChange?.(event as React.ChangeEvent<HTMLInputElement>);
        } else if (event.type === "input") {
          onInput?.(event as React.FormEvent<HTMLInputElement>);
        }
      }
    },
    [onChange, onInput]
  );

  const cacheSelection = React.useCallback(
    (target: HTMLInputElement | null) => {
      if (!target) {
        return;
      }

      selectionRef.current = {
        start: target.selectionStart ?? actualValueRef.current.length,
        end: target.selectionEnd ?? actualValueRef.current.length,
      };
    },
    []
  );

  const handleChange = React.useCallback(
    (event: React.ChangeEvent<HTMLInputElement>) => {
      const nativeEvent = event.nativeEvent as InputEvent;
      const data = nativeEvent.data ?? "";
      const inputType = nativeEvent.inputType ?? "insertText";
      const { start, end } = selectionRef.current;
      let nextValue = actualValueRef.current;
      let caretPosition = start;

      if (inputType.startsWith("delete")) {
        const rangeLength = end - start;
        const deleteCount =
          rangeLength > 0
            ? rangeLength
            : inputType === "deleteContentBackward" ||
              inputType === "deleteByCut"
            ? 1
            : inputType === "deleteContentForward"
            ? 1
            : 0;
        const deleteStart =
          rangeLength > 0
            ? start
            : inputType === "deleteContentBackward"
            ? Math.max(start - 1, 0)
            : start;
        nextValue =
          nextValue.slice(0, deleteStart) +
          nextValue.slice(deleteStart + deleteCount);
        caretPosition = deleteStart;
      } else {
        nextValue = nextValue.slice(0, start) + data + nextValue.slice(end);
        caretPosition = start + data.length;
      }

      actualValueRef.current = nextValue;
      setMaskedValue(maskValue(nextValue));
      selectionRef.current = {
        start: caretPosition,
        end: caretPosition,
      };

      emitChange(nextValue, event);

      requestAnimationFrame(() => {
        if (inputRef.current) {
          inputRef.current.selectionStart = caretPosition;
          inputRef.current.selectionEnd = caretPosition;
        }
      });
    },
    [emitChange]
  );

  const handleInput = React.useCallback(
    (event: React.FormEvent<HTMLInputElement>) => {
      emitChange(actualValueRef.current, event);
    },
    [emitChange]
  );

  return {
    inputRef,
    maskedValue,
    cacheSelection,
    handleChange,
    handleInput,
  };
}

const BaseInput = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, type = "text", ...props }, ref) => (
    <input
      ref={ref}
      type={type}
      data-slot="input"
      className={cn(
        "file:text-foreground placeholder:text-muted-foreground selection:bg-primary selection:text-primary-foreground dark:bg-input/30 border-input h-9 w-full min-w-0 rounded-md border bg-transparent px-3 py-1 text-base shadow-xs transition-[color,box-shadow] outline-none file:inline-flex file:h-7 file:border-0 file:bg-transparent file:text-sm file:font-medium disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-50 md:text-sm",
        "focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px]",
        "aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive",
        className
      )}
      {...props}
    />
  )
);
BaseInput.displayName = "BaseInput";

const PasswordInput = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, value, defaultValue, onChange, onInput, ...props }, ref) => {
    const { inputRef, maskedValue, cacheSelection, handleChange, handleInput } =
      useMaskedPassword({ value, defaultValue, onChange, onInput });
    React.useImperativeHandle(ref, () => inputRef.current as HTMLInputElement);

    return (
      <BaseInput
        {...props}
        ref={inputRef}
        className={className}
        type="text"
        autoComplete={props.autoComplete ?? "new-password"}
        value={maskedValue}
        onBeforeInput={(event) => cacheSelection(event.currentTarget)}
        onKeyDown={(event) => cacheSelection(event.currentTarget)}
        onSelect={(event) => cacheSelection(event.currentTarget)}
        onFocus={(event) => cacheSelection(event.currentTarget)}
        onChange={handleChange}
        onInput={handleInput}
      />
    );
  }
);
PasswordInput.displayName = "PasswordInput";

const Input = React.forwardRef<HTMLInputElement, InputProps>((props, ref) => {
  if (props.type === "password") {
    return <PasswordInput {...props} ref={ref} />;
  }

  return <BaseInput {...props} ref={ref} />;
});
Input.displayName = "Input";

export { Input };
