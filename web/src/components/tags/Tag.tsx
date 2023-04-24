import classNames from "classnames";
import styles from "./tag.module.scss";
import { LockClosed16Regular as LockedIcon } from "@fluentui/react-icons";

export enum TagSize {
  normal = "normal",
  small = "small",
  tiny = "tiny",
}

export const TagSizeClasses = new Map<TagSize, string>([
  [TagSize.normal, styles.normal],
  [TagSize.small, styles.small],
  [TagSize.tiny, styles.tiny],
]);

export enum TagColor {
  primary = "primary",
  success = "success",
  warning = "warning",
  error = "error",
  accent1 = "accent1",
  accent2 = "accent2",
  accent3 = "accent3",
  custom = "custom",
}

export const TagColorClasses = new Map<TagColor, string>([
  [TagColor.primary, styles.primary],
  [TagColor.success, styles.success],
  [TagColor.warning, styles.warning],
  [TagColor.error, styles.error],
  [TagColor.accent1, styles.accent1],
  [TagColor.accent2, styles.accent2],
  [TagColor.accent3, styles.accent3],
  [TagColor.custom, styles.custom],
]);

export type TagProps = {
  sizeVariant?: TagSize;
  colorVariant?: TagColor;
  customBackgroundColor?: string;
  customTextColor?: string;
  locked?: boolean;
  label: string;
};

export default function Tag(props: TagProps) {
  const sizeClass = TagSizeClasses.get(props.sizeVariant || TagSize.small);
  const colorClass = TagColorClasses.get(props.colorVariant || TagColor.primary);
  const customColor = {
    backgroundColor: props.customBackgroundColor,
    color: props.customTextColor,
  };
  return (
    <div className={classNames(styles.tagWrapper, colorClass, sizeClass)} style={customColor}>
      {props.locked && (
        <div className={classNames(styles.icon)}>
          <LockedIcon />
        </div>
      )}
      <div className={styles.tagLabel}>{props.label}</div>
    </div>
  );
}
