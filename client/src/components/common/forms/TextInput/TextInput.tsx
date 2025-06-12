import { Component } from "solid-js";
import styles from "./TextInput.module.scss";

type TextInputType = "text" | "password" | "email";

interface TextInputProps {
  label?: string;
  defaultValue?: string;
  onBlur?: (
    value: string,
    target: HTMLInputElement,
    event: FocusEvent & { target: HTMLInputElement },
  ) => void;
  type?: TextInputType;
  autoComplete?: string;
}

export const TextInput: Component<TextInputProps> = (props) => {
  const handleUpdate = (event: FocusEvent & { target: HTMLInputElement }) => {
    props.onBlur?.(event.target.value, event.target, event);
  };

  return (
    <div class={styles.textInput}>
      <label class={styles.label}>
        <span class={styles.labelText}>{props.label}</span>
      </label>
      <input
        type={props.type ?? "text"}
        class={styles.input}
        value={props.defaultValue ?? ""}
        onBlur={handleUpdate}
        autocomplete={props.autoComplete ?? "off"}
      />
    </div>
  );
};
