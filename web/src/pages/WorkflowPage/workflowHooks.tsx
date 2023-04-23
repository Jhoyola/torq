import { Add16Regular as NewWorkflowIcon } from "@fluentui/react-icons";
import Button, { ColorVariant } from "components/buttons/Button";
import useTranslations from "services/i18n/useTranslations";
import { useNavigate } from "react-router";
import { useGetWorkflowQuery, useNewWorkflowMutation } from "pages/WorkflowPage/workflowApi";
import { ReactNode } from "react";
import { Workflow, WorkflowVersion, WorkflowVersionNode } from "./workflowTypes";
import { userEvents } from "utils/userEvents";

export function useNewWorkflowButton(): ReactNode {
  const { t } = useTranslations();
  const { track } = userEvents();
  const navigate = useNavigate();
  const [newWorkflow] = useNewWorkflowMutation();

  function newWorkflowHandler() {
    const response = newWorkflow();
    track("Workflow Create");
    response
      .then((res) => {
        const data = (res as { data: { workflowId: number; version: number } }).data;
        track("Navigate to Workflow", {
          workflowId: data.workflowId,
          workflowVersion: data.version,
        });
        navigate(`/manage/workflows/${data.workflowId}/versions/${data.version}`);
      })
      .catch((err) => {
        // TODO: Handle error and show a toast
        console.log(err);
      });
  }

  return (
    <Button
      intercomTarget={"new-workflow-button"}
      buttonColor={ColorVariant.success}
      hideMobileText={true}
      icon={<NewWorkflowIcon />}
      onClick={newWorkflowHandler}
    >
      {t.newWorkflow}
    </Button>
  );
}

export function useWorkflowData(workflowId?: string, version?: string) {
  const { data } = useGetWorkflowQuery(
    {
      workflowId: parseInt(workflowId || ""),
      version: parseInt(version || ""),
    },
    { skip: !workflowId || !version }
  );

  const workflow: Workflow | undefined = data?.workflow;
  const workflowVersion: WorkflowVersion | undefined = data?.version;

  // reduce the workflow nodes to an object of stages containing an array of nodes
  const stageNumbers: Array<number> = (
    (data?.nodes || []).reduce((acc: Array<number>, node: WorkflowVersionNode) => {
      if (node.stage && !acc.includes(node.stage)) {
        acc.push(node.stage);
      }
      return acc;
    }, []) || []
  ).sort((a, b) => a - b);

  return { workflow, workflowVersion, stageNumbers };
}
