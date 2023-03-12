import {
  ArrowRouting20Regular as ChannelsIcon,
  ArrowSyncFilled as ProcessingIcon,
  Checkmark20Regular as SuccessNoteIcon,
  CheckmarkRegular as SuccessIcon,
  ErrorCircleRegular as FailedIcon,
  Link20Regular as LinkIcon,
  Note20Regular as NoteIcon,
} from "@fluentui/react-icons";
import { ChangeEvent, useState } from "react";
import Button, { ButtonWrapper, ColorVariant, ExternalLinkButton } from "components/buttons/Button";
import ProgressHeader, { ProgressStepState, Step } from "features/progressTabs/ProgressHeader";
import ProgressTabs, { ProgressTabContainer } from "features/progressTabs/ProgressTab";
import styles from "./closeChannel.module.scss";
import { useNavigate } from "react-router";
import PopoutPageTemplate from "features/templates/popoutPageTemplate/PopoutPageTemplate";
import useTranslations from "services/i18n/useTranslations";
import classNames from "classnames";
import { NumberFormatValues } from "react-number-format";
import Input from "components/forms/input/Input";
import Switch from "components/forms/switch/Switch";
import FormRow from "features/forms/FormWrappers";
import { useSearchParams } from "react-router-dom";
import Note, { NoteType } from "features/note/Note";
import mixpanel from "mixpanel-browser";
import { useCloseChannelMutation } from "./closeChannelApi";
import ErrorSummary from "components/errors/ErrorSummary";
import { RtqToServerError } from "components/errors/errors";

const closeStatusClass = {
  IN_FLIGHT: styles.inFlight,
  FAILED: styles.failed,
  SUCCEEDED: styles.success,
};

const closeStatusIcon = {
  IN_FLIGHT: <ProcessingIcon />,
  FAILED: <FailedIcon />,
  SUCCEEDED: <SuccessIcon />,
  NOTE: <NoteIcon />,
};

function closeChannelModal() {
  const { t } = useTranslations();
  const navigate = useNavigate();
  const [queryParams] = useSearchParams();
  const nodeId = parseInt(queryParams.get("nodeId") || "0");
  const channelId = parseInt(queryParams.get("channelId") || "0");

  const [resultState, setResultState] = useState(ProgressStepState.disabled);
  const [detailState, setDetailState] = useState(ProgressStepState.active);
  const [satPerVbyte, setSatPerVbyte] = useState<number | undefined>();
  const [stepIndex, setStepIndex] = useState(0);
  const [force, setForce] = useState<boolean>(false);

  const closeAndReset = () => {
    setStepIndex(0);
    setDetailState(ProgressStepState.active);
    setResultState(ProgressStepState.disabled);
  };

  const [closeChannel, { data: closeChannelResponse, error: closeChannelError, isError }] = useCloseChannelMutation();

  return (
    <PopoutPageTemplate title={"Close Channel"} show={true} onClose={() => navigate(-1)} icon={<ChannelsIcon />}>
      <ProgressHeader modalCloseHandler={closeAndReset}>
        <Step label={"Detail"} state={detailState} last={false} />
        <Step label={"Result"} state={resultState} last={true} />
      </ProgressHeader>

      <ProgressTabs showTabIndex={stepIndex}>
        <ProgressTabContainer>
          <div className={styles.activeColumns}>
            <div className={styles.closeChannelTableRow}>
              <FormRow>
                <div className={styles.closeChannelTableSingle}>
                  <span className={styles.label}>{"Sat per vbyte"}</span>
                  <div className={styles.input}>
                    <Input
                      formatted={true}
                      className={styles.single}
                      thousandSeparator={","}
                      value={satPerVbyte}
                      suffix={" sat/vbyte"}
                      onValueChange={(values: NumberFormatValues) => {
                        setSatPerVbyte(values.floatValue);
                      }}
                    />
                  </div>
                </div>
              </FormRow>
            </div>
            {/*<div className={styles.closeChannelTableRow}>*/}
            {/*  <FormRow>*/}
            {/*    <div className={styles.closeChannelTableSingle}>*/}
            {/*      <span className={styles.label}>{"Close Address (for local funds)"}</span>*/}
            {/*      <div className={styles.input}>*/}
            {/*        <Input*/}
            {/*          type={"text"}*/}
            {/*          value={closeAddress}*/}
            {/*          placeholder={"e.g. bc1q..."}*/}
            {/*          onChange={(e: ChangeEvent<HTMLInputElement>) => {*/}
            {/*            setCloseAddress(e.target.value);*/}
            {/*          }}*/}
            {/*        />*/}
            {/*      </div>*/}
            {/*    </div>*/}
            {/*  </FormRow>*/}
            {/*</div>*/}
            <div className={styles.closeChannelTableRow}>
              <FormRow className={styles.switchRow}>
                <Switch
                  label={"Force close"}
                  checked={force}
                  onChange={(e: ChangeEvent<HTMLInputElement>) => {
                    setForce(e.target.checked);
                  }}
                />
              </FormRow>
            </div>
            <ButtonWrapper
              rightChildren={
                <Button
                  onClick={() => {
                    setStepIndex(1);
                    setDetailState(ProgressStepState.completed);
                    setResultState(ProgressStepState.completed);
                    mixpanel.track("Close Channel", {
                      nodeId: nodeId,
                      channelId: channelId,
                      openChannelUseSatPerVbyte: satPerVbyte !== 0,
                      force: force,
                    });
                    closeChannel({
                      nodeId: nodeId,
                      channelId: channelId,
                      satPerVbyte: satPerVbyte,
                      force: force,
                    });
                  }}
                  buttonColor={ColorVariant.success}
                >
                  {t.openCloseChannel.closeChannel}
                </Button>
              }
            />
          </div>
        </ProgressTabContainer>
        <ProgressTabContainer>
          <div
            className={classNames(
              styles.closeChannelResultIconWrapper,
              { [styles.failed]: isError },
              closeStatusClass[isError ? "FAILED" : "SUCCEEDED"]
            )}
          >
            {" "}
            {closeStatusIcon[isError ? "FAILED" : "SUCCEEDED"]}
          </div>
          <div className={styles.closeChannelResultDetails}>
            {!isError && (
              <>
                <Note title={t.TxId} icon={<SuccessNoteIcon />} noteType={NoteType.success}>
                  {closeChannelResponse?.closingTransactionHash}
                </Note>
                <ExternalLinkButton
                  href={"https://mempool.space/tx/" + closeChannelResponse?.closingTransactionHash}
                  target="_blank"
                  rel="noreferrer"
                  buttonColor={ColorVariant.success}
                  icon={<LinkIcon />}
                >
                  {t.openCloseChannel.GoToMempool}
                </ExternalLinkButton>

                <Note title={t.note} icon={<NoteIcon />} noteType={NoteType.info}>
                  {t.openCloseChannel.confirmationClosing}
                </Note>
              </>
            )}
            <ErrorSummary title={t.Error} errors={RtqToServerError(closeChannelError).errors} />
          </div>
        </ProgressTabContainer>
      </ProgressTabs>
    </PopoutPageTemplate>
  );
}

export default closeChannelModal;
