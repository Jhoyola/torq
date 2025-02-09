import {
  ArrowSyncFilled as ProcessingIcon,
  CheckmarkRegular as SuccessIcon,
  ErrorCircleRegular as FailedIcon,
  Note20Regular as NoteIcon,
  Molecule20Regular as PeersIcon,
} from "@fluentui/react-icons";
import { useEffect, useState } from "react";
import Button, { ButtonWrapper, ColorVariant } from "components/buttons/Button";
import ProgressHeader, { ProgressStepState, Step } from "features/progressTabs/ProgressHeader";
import ProgressTabs, { ProgressTabContainer } from "features/progressTabs/ProgressTab";
import styles from "./peers.module.scss";
import { useNavigate } from "react-router";
import PopoutPageTemplate from "features/templates/popoutPageTemplate/PopoutPageTemplate";
import useTranslations from "services/i18n/useTranslations";
import classNames from "classnames";
import FormRow from "features/forms/FormWrappers";
import Note, { NoteType } from "features/note/Note";
import { userEvents } from "utils/userEvents";
import { FormErrors, mergeServerError, ServerErrorType } from "components/errors/errors";
import ErrorSummary from "components/errors/ErrorSummary";
import { useGetPeersQuery, useUpdatePeerMutation } from "./peersApi";
import clone from "clone";
import { useSearchParams } from "react-router-dom";
import { NodeConnectionSetting, Peer } from "./peersTypes";
import { useAppSelector } from "store/hooks";
import { selectActiveNetwork } from "../network/networkSlice";

const updateStatusClass = {
  PROCESSING: styles.processing,
  FAILED: styles.failed,
  SUCCEEDED: styles.success,
};

const updateStatusIcon = {
  PROCESSING: <ProcessingIcon />,
  FAILED: <FailedIcon />,
  SUCCEEDED: <SuccessIcon />,
  NOTE: <NoteIcon />,
};

function PeerUpdateModal() {
  const { track } = userEvents();
  const [queryParams] = useSearchParams();
  const peerNodeId = parseInt(queryParams.get("peerNodeId") || "0");
  const torqNodeId = parseInt(queryParams.get("torqNodeId") || "0");
  const activeNetwork = useAppSelector(selectActiveNetwork);

  const peersResponse = useGetPeersQuery<{
    data: Array<Peer>;
    isLoading: boolean;
    isFetching: boolean;
    isUninitialized: boolean;
    isSuccess: boolean;
  }>({ network: activeNetwork });

  const peer = peersResponse?.data?.find((peer: Peer) => peer.nodeId === peerNodeId && peer.torqNodeId === torqNodeId);

  const { t } = useTranslations();

  const navigate = useNavigate();

  const [connectState, setConnectState] = useState(ProgressStepState.active);
  const [stepIndex, setStepIndex] = useState(0);

  const [selectedSetting, setSelectedSetting] = useState<NodeConnectionSetting | undefined>();
  const [resultState, setResultState] = useState(ProgressStepState.disabled);
  const [formErrorState, setFormErrorState] = useState({} as FormErrors);
  const [updatePeer, response] = useUpdatePeerMutation();

  useEffect(() => {
    if (peer) {
      setSelectedSetting(peer.setting);
    }
  }, [peer]);

  useEffect(() => {
    if (response && response.isError && response.error && "data" in response.error && response.error.data) {
      const mergedErrors = mergeServerError(response.error.data as ServerErrorType, clone(formErrorState));
      setFormErrorState(mergedErrors);
      setResultState(ProgressStepState.error);
    }
    if (response && response.isLoading) {
      setResultState(ProgressStepState.processing);
    }
    if (response.isSuccess) {
      setResultState(ProgressStepState.completed);
    }
  }, [response]);

  const closeAndReset = () => {
    setStepIndex(0);
    setConnectState(ProgressStepState.active);

    setResultState(ProgressStepState.disabled);
  };

  function handleConnectPeer() {
    if (selectedSetting === undefined) return;

    track("Update Peer", {
      peerNodeId: peerNodeId,
      torqNodeId: torqNodeId,
    });
    updatePeer({
      nodeId: peerNodeId,
      torqNodeId: torqNodeId,
    });
  }

  return (
    <PopoutPageTemplate title={t.peersPage.updatePeer} show={true} onClose={() => navigate(-1)} icon={<PeersIcon />}>
      <ProgressHeader modalCloseHandler={closeAndReset}>
        <Step label={"Update"} state={connectState} last={false} />
        <Step label={"Result"} state={resultState} last={true} />
      </ProgressHeader>

      <ProgressTabs showTabIndex={stepIndex}>
        <ProgressTabContainer>
          <FormRow>
            <div className={styles.card}>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Peer alias</div>
                <div className={styles.rowValue}>{peer?.peerAlias}</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Torq node name</div>
                <div className={classNames(styles.rowValue)}>{peer?.nodeName}</div>
              </div>
            </div>
          </FormRow>
          <ButtonWrapper
            rightChildren={
              <Button
                intercomTarget={"update-peer-confirm-button"}
                onClick={() => {
                  setStepIndex(1);
                  setConnectState(ProgressStepState.completed);
                  handleConnectPeer();
                }}
                buttonColor={ColorVariant.success}
              >
                {t.update}
              </Button>
            }
          />
        </ProgressTabContainer>
        <ProgressTabContainer>
          <div
            className={classNames(
              styles.peerResultIconWrapper,
              { [styles.failed]: response.isError },
              updateStatusClass[response.isLoading ? "PROCESSING" : response.isError ? "FAILED" : "SUCCEEDED"]
            )}
          >
            {updateStatusIcon[response.isLoading ? "PROCESSING" : response.isSuccess ? "SUCCEEDED" : "FAILED"]}
          </div>
          {response.isLoading && (
            <Note title={t.Processing} icon={<ProcessingIcon />} noteType={NoteType.warning}>
              {t.openCloseChannel.processingClose}
            </Note>
          )}
          <div className={styles.peersResultDetails}>
            {response.isSuccess && (
              <Note title={t.Success} icon={<NoteIcon />} noteType={NoteType.success}>
                {t.peersPage.confirmationPeerUpdated}
              </Note>
            )}
            <ErrorSummary errors={formErrorState} />
            <ButtonWrapper
              className={styles.resetButton}
              rightChildren={
                <Button
                  intercomTarget={"update-peer-close-button"}
                  onClick={() => {
                    closeAndReset();
                  }}
                  buttonColor={ColorVariant.primary}
                >
                  {t.peersPage.close}
                </Button>
              }
            />
          </div>
        </ProgressTabContainer>
      </ProgressTabs>
    </PopoutPageTemplate>
  );
}

export default PeerUpdateModal;
