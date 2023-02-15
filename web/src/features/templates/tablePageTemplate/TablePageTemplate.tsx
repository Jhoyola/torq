import React from "react";
import styles from "./table-page-template.module.scss";

import classNames from "classnames";
import { FluentIconsProps } from "@fluentui/react-icons";
import PageTitle from "features/templates/PageTitle";
import Button, { ColorVariant } from "components/buttons/Button";

type TablePageTemplateProps = {
  title: string;
  titleContent?: React.ReactNode;
  sidebarExpanded?: boolean;
  sidebar?: React.ReactNode;
  pagination?: React.ReactNode;
  pageTotals?: React.ReactNode;
  tableControls?: React.ReactNode;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  breadcrumbs?: Array<any>;
  children?: React.ReactNode;
  onNameChange?: (title: string) => void;
};

export default function TablePageTemplate(props: TablePageTemplateProps) {
  return (
    <div className={classNames(styles.contentWrapper)}>
      <PageTitle breadcrumbs={props.breadcrumbs} title={props.title} onNameChange={props.onNameChange}>
        {props.titleContent}
      </PageTitle>

      {props.tableControls}

      <div className={styles.tableWrapper}>
        <div className={styles.tableContainer}>
          <div className={styles.tableExpander}>{props.children}</div>
        </div>
      </div>

      <div className={classNames(styles.paginationWrapper)}>{props.pagination}</div>

      <div className={classNames(styles.pageSidebarWrapper, { [styles.sidebarExpanded]: props.sidebarExpanded })}>
        {props.sidebar}
      </div>
    </div>
  );
}

type TableControlsButtonProps = {
  id?: string;
  text?: string;
  icon: React.FC<FluentIconsProps>;
  onClickHandler?: () => void;
  active?: boolean;
};

export function TableControlsButton(props: TableControlsButtonProps) {
  return (
    <div className={classNames(styles.tableControlsButtonWrapper, { [styles.active]: props.active })}>
      <Button
        id={props.id}
        className={styles.tableControlsButtonIcon}
        onClick={props.onClickHandler}
        buttonColor={ColorVariant.primary}
        icon={<props.icon />}
      >
        {props.text}
      </Button>
      {/*<div className={styles.title}>{props.title}</div>*/}
    </div>
  );
}

export function TableControlSection(props: { children?: React.ReactNode }) {
  return <div className={classNames(styles.tableControlsSection)}>{props.children}</div>;
}

export function TableControlsButtonGroup(props: { children?: React.ReactNode }) {
  return <div className={classNames(styles.tableControlsButtonGroup)}>{props.children}</div>;
}

export function TableControlsTabsGroup(props: { children?: React.ReactNode }) {
  return <div className={classNames(styles.tableControlsTabsGroup)}>{props.children}</div>;
}
