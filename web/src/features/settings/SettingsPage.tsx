import { AddSquare20Regular as AddIcon, Save20Regular as SaveIcon } from "@fluentui/react-icons";
import { torqApi } from "apiSlice";
import { useAppDispatch } from "store/hooks";
import Page from "layout/Page";
import Button, { ColorVariant, ButtonPosition } from "components/buttons/Button";
import styles from "features/settings/settings.module.css";
import { SelectOption } from "features/forms/Select";
import Select from "components/forms/select/Select";
import React from "react";
import { defaultStaticRangesFn } from "features/timeIntervalSelect/customRanges";
import {
  useGetNodeConfigurationsQuery,
  useGetSettingsQuery,
  useGetTimeZonesQuery,
  useUpdateSettingsMutation,
} from "apiSlice";
import { settings } from "apiTypes";
import { toastCategory } from "features/toast/Toasts";
import ToastContext from "features/toast/context";
import NodeSettings from "features/settings/NodeSettings";
import Modal from "features/modal/Modal";
import { useGetServicesQuery } from "apiSlice";
import useTranslations from "services/i18n/useTranslations";
import { supportedLangs } from "config/i18nConfig";
import Input from "components/formsWithValidation/input/InputWithValidation";

function Settings() {
  const { t, setLang } = useTranslations();
  const { data: settingsData } = useGetSettingsQuery();
  const { data: nodeConfigurations } = useGetNodeConfigurationsQuery();
  const { data: timeZones = [] } = useGetTimeZonesQuery();
  const [updateSettings] = useUpdateSettingsMutation();
  const toastRef = React.useContext(ToastContext);
  const addNodeRef = React.useRef(null);

  const [showAddNodeState, setShowAddNodeState] = React.useState(false);
  const [settingsState, setSettingsState] = React.useState({} as settings);
  const dispatch = useAppDispatch();

  React.useEffect(() => {
    if (settingsData) {
      setSettingsState(settingsData);
    }
  }, [settingsData]);

  const defaultDateRangeLabels: {
    label: string;
    code: string;
  }[] = defaultStaticRangesFn(0);

  const defaultDateRangeOptions: SelectOption[] = defaultDateRangeLabels.map((dsr) => ({
    value: dsr.code,
    label: dsr.label,
  }));

  const preferredTimezoneOptions: SelectOption[] = timeZones.map((tz) => ({
    value: tz.name,
    label: tz.name,
  }));

  const weekStartsOnOptions: SelectOption[] = [
    { label: t.saturday, value: "saturday" },
    { label: t.sunday, value: "sunday" },
    { label: t.monday, value: "monday" },
  ];

  // When adding a language also add it to web/src/config/i18nConfig.js
  const languageOptions: SelectOption[] = [
    { label: supportedLangs.en, value: "en" },
    { label: supportedLangs.nl, value: "nl" },
  ];

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const handleDefaultDateRangeChange = (combiner: any) => {
    setSettingsState({ ...settingsState, defaultDateRange: combiner.value });
  };

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const handleDefaultLanguageRangeChange = (combiner: any) => {
    setSettingsState({ ...settingsState, defaultLanguage: combiner.value });
  };

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const handlePreferredTimezoneChange = (combiner: any) => {
    setSettingsState({
      ...settingsState,
      preferredTimezone: combiner.value,
    });
  };

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const handleWeekStartsOnChange = (combiner: any) => {
    setSettingsState({ ...settingsState, weekStartsOn: combiner.value });
  };

  const handleSlackOAuthTokenChange = (value: string) => {
    setSettingsState({ ...settingsState, slackOAuthToken: value });
  };

  const handleSlackBotAppTokenChange = (value: string) => {
    setSettingsState({ ...settingsState, slackBotAppToken: value });
  };

  const handleTelegramHighPriorityCredentialsChange = (value: string) => {
    setSettingsState({ ...settingsState, telegramHighPriorityCredentials: value });
  };

  const handleTelegramLowPriorityCredentialsChange = (value: string) => {
    setSettingsState({ ...settingsState, telegramLowPriorityCredentials: value });
  };

  const submitPreferences = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    updateSettings(settingsState);
    setLang(settingsState?.defaultLanguage);
    dispatch(torqApi.util.resetApiState());
    toastRef?.current?.addToast(t.toast.settingsSaved, toastCategory.success);
  };

  const addNodeConfiguration = () => {
    setShowAddNodeState(true);
  };

  const handleNewNodeModalOnClose = () => {
    if (addNodeRef.current) {
      (addNodeRef.current as { clear: () => void }).clear();
    }
    setShowAddNodeState(false);
  };

  const handleOnAddSuccess = () => {
    setShowAddNodeState(false);
  };

  // We only fetch the Version once when intial rendering.
  const { data: servicesData } = useGetServicesQuery();
  const [version = "Unknown", commit = "Unknown"] = servicesData?.version?.split(" | ") ?? [];

  return (
    <Page>
      <React.Fragment>
        <div className={styles.settingsPage}>
          <div className={styles.settingsColumn}>
            <div data-intercom-target="settings-general-settings-card">
              <h3>{t.settings}</h3>

              <form onSubmit={submitPreferences} className={styles.settingsForm}>
                <Select
                  intercomTarget="settings-default-date-range"
                  label={t.defaultDateRange}
                  onChange={handleDefaultDateRangeChange}
                  options={defaultDateRangeOptions}
                  value={defaultDateRangeOptions.find((dd) => dd.value === settingsState?.defaultDateRange)}
                />
                <Select
                  intercomTarget="settings-default-languag"
                  label={t.language}
                  onChange={handleDefaultLanguageRangeChange}
                  options={languageOptions}
                  value={languageOptions.find((lo) => lo.value === settingsState?.defaultLanguage)}
                />
                <Select
                  intercomTarget="settings-preferred-timezone"
                  label={t.preferredTimezone}
                  onChange={handlePreferredTimezoneChange}
                  options={preferredTimezoneOptions}
                  value={preferredTimezoneOptions.find((tz) => tz.value === settingsState?.preferredTimezone)}
                />
                <Select
                  intercomTarget="settings-week-starts-on"
                  label={t.weekStartsOn}
                  onChange={handleWeekStartsOnChange}
                  options={weekStartsOnOptions}
                  value={weekStartsOnOptions.find((dd) => dd.value === settingsState?.weekStartsOn)}
                />
                <div data-intercom-target={"settings-slack-section"}>
                  <Input
                    intercomTarget="settings-slack-oauth-token"
                    label={t.slackOAuthToken}
                    value={settingsState?.slackOAuthToken}
                    type={"text"}
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => handleSlackOAuthTokenChange(e.target.value)}
                  />
                  <Input
                    intercomTarget="settings-slack-bot-app-token"
                    label={t.slackBotAppToken}
                    value={settingsState?.slackBotAppToken}
                    type={"text"}
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => handleSlackBotAppTokenChange(e.target.value)}
                  />
                </div>
                <div data-intercom-target={"settings-telegram-section"}>
                  <Input
                    intercomTarget="settings-telegram-high-priority-credentials"
                    label={t.telegramHighPriorityCredentials}
                    value={settingsState?.telegramHighPriorityCredentials}
                    type={"text"}
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                      handleTelegramHighPriorityCredentialsChange(e.target.value)
                    }
                  />
                  <Input
                    intercomTarget="settings-telegram-low-priority-credentials"
                    label={t.telegramLowPriorityCredentials}
                    value={settingsState?.telegramLowPriorityCredentials}
                    type={"text"}
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                      handleTelegramLowPriorityCredentialsChange(e.target.value)
                    }
                  />
                </div>
                <Button
                  intercomTarget="settings-save-button"
                  type={"submit"}
                  icon={<SaveIcon />}
                  buttonColor={ColorVariant.success}
                  buttonPosition={ButtonPosition.fullWidth}
                >
                  {t.save}
                </Button>
              </form>
            </div>
            <div className={styles.nodeSettingsWrapper} data-intercom-target="settings-node-settings-card">
              <h3>{t.header.nodes}</h3>
              {nodeConfigurations &&
                nodeConfigurations?.map((nodeConfiguration) => (
                  <NodeSettings
                    nodeId={nodeConfiguration.nodeId}
                    key={nodeConfiguration.nodeId ?? 0}
                    collapsed={true}
                  />
                ))}
              <Button
                buttonColor={ColorVariant.success}
                onClick={addNodeConfiguration}
                icon={<AddIcon />}
                intercomTarget={"settings-add-node-button"}
              >
                {t.addNode}
              </Button>
            </div>
            <Modal title={t.addNode} show={showAddNodeState} onClose={handleNewNodeModalOnClose}>
              <NodeSettings
                ref={addNodeRef}
                addMode={true}
                nodeId={0}
                collapsed={false}
                onAddSuccess={handleOnAddSuccess}
              />
            </Modal>
            <footer className={styles.footer}>
              © Torq: {version} | Commit: {commit}
            </footer>
          </div>
        </div>
      </React.Fragment>
    </Page>
  );
}

export default Settings;
