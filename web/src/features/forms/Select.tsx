import RawSelect, { SelectOptionType } from "components/forms/select/Select";
import { Props } from "react-select";

type selectProps = {
  label: string;
  intercomTarget: string;
} & Props;

export type SelectOptions = {
  label?: string;
  value: number | string;
  type?: string;
};

function Select(props: selectProps) {
  return <RawSelect intercomTarget={props.intercomTarget} label={props.label} options={props.options} value={props.value} onChange={props.onChange} />;
}

export default Select;
export type SelectOption = SelectOptionType;
