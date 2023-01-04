import { useAppDispatch, useAppSelector } from "store/hooks";
import { Globe20Regular as GlobeIcon } from "@fluentui/react-icons";
import styles from "./nav.module.scss";
import Popover from "features/popover/Popover";
import Button, { ColorVariant, SizeVariant } from "components/buttons/Button";
import { selectActiveNetwork, setActiveNetwork, Network } from "features/network/networkSlice";
import useTranslations from "services/i18n/useTranslations";

function QuickToggles() {
  const { t } = useTranslations();
  const dispatch = useAppDispatch();
  const activeNetwork = useAppSelector(selectActiveNetwork);

  return (
    <div className={styles.quickToggles}>
      <Popover
        button={
          <Button
            buttonColor={ColorVariant.ghost}
            buttonSize={SizeVariant.small}
            isOpen={false}
            icon={<GlobeIcon />}
            className={"collapse-tablet"}
          />
        }
        className={"right"}
      >
        <div className={styles.quickToggleContent}>
          <Button
            buttonColor={activeNetwork === Network.MainNet ? ColorVariant.primary : ColorVariant.primary}
            onClick={() => dispatch(setActiveNetwork(Network.MainNet))}
          >
            {t.MainNet}
          </Button>
          <Button
            buttonColor={activeNetwork === Network.TestNet ? ColorVariant.primary : ColorVariant.primary}
            onClick={() => dispatch(setActiveNetwork(Network.TestNet))}
          >
            {t.TestNet}
          </Button>
          <Button
            buttonColor={activeNetwork === Network.RegTest ? ColorVariant.primary : ColorVariant.primary}
            onClick={() => dispatch(setActiveNetwork(Network.RegTest))}
          >
            {t.RegTest}
          </Button>
          <Button
            buttonColor={activeNetwork === Network.SigNet ? ColorVariant.primary : ColorVariant.primary}
            onClick={() => dispatch(setActiveNetwork(Network.SigNet))}
          >
            {t.SigNet}
          </Button>
          <Button
            buttonColor={activeNetwork === Network.SimNet ? ColorVariant.primary : ColorVariant.primary}
            onClick={() => dispatch(setActiveNetwork(Network.SimNet))}
          >
            {t.SimNet}
          </Button>
        </div>
      </Popover>
    </div>
  );
}

export default QuickToggles;
