import { Component, JSX } from "solid-js";
import styles from "./Button.module.scss";

export type ButtonVariant = "primary" | "secondary" | "tertiary" | "danger";
export type ButtonSize = "sm" | "md" | "lg";

interface ButtonProps {
  variant?: ButtonVariant;
  size?: ButtonSize;
  type?: "button" | "submit" | "reset";
  disabled?: boolean;
  onClick?: (event: MouseEvent) => void;
  children: JSX.Element;
}

export const Button: Component<ButtonProps> = (props) => {
  const variant = props.variant || "primary";
  const size = props.size || "md";

  return (
    <button
      class={`${styles.button} ${styles[variant]} ${styles[size]}`}
      type={props.type || "button"}
      disabled={props.disabled}
      onClick={props.onClick}
    >
      {props.children}
    </button>
  );
};
