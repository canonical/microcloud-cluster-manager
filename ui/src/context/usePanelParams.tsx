import { useSearchParams } from "react-router-dom";

export interface PanelHelper {
  panel: string | null;
  cluster: string | null;
  clusters: string | null;
  focusField: string | null;
  clear: () => void;
  openEnrolCluster: () => void;
  openConfigureCluster: (cluster?: string, focusField?: string) => void;
  openBulkConfigureCluster: (clusterNames: string[]) => void;
}

export const panels = {
  enrolCluster: "enrol-cluster",
  configureCluster: "configure-cluster",
  bulkConfigureCluster: "configure-cluster-bulk",
};

type ParamMap = Record<string, string>;

const usePanelParams = (): PanelHelper => {
  const [params, setParams] = useSearchParams();

  const craftResizeEvent = () => {
    setTimeout(() => window.dispatchEvent(new Event("resize")), 100);
  };

  const setPanelParams = (panel: string, args: ParamMap = {}) => {
    const newParams = new URLSearchParams();
    newParams.set("panel", panel);
    for (const [key, value] of Object.entries(args)) {
      if (value) {
        newParams.set(key, value);
      }
    }
    setParams(newParams);
    craftResizeEvent();
  };

  const clearParams = () => {
    const newParams = new URLSearchParams(params);
    // we only want to remove search params set when opening the panel
    // pre-existing search params should be kept e.g. params from the search bar
    newParams.delete("panel");
    newParams.delete("cluster");
    newParams.delete("clusters");
    newParams.delete("focusField");
    setParams(newParams);
    craftResizeEvent();
  };

  return {
    panel: params.get("panel"),
    cluster: params.get("cluster"),
    clusters: params.get("clusters"),
    focusField: params.get("focusField"),

    clear: () => {
      clearParams();
    },

    openEnrolCluster: () => {
      setPanelParams(panels.enrolCluster);
    },

    openConfigureCluster: (cluster?: string, focusField?: string) => {
      const params = { cluster: cluster || "", focusField: focusField || "" };
      setPanelParams(panels.configureCluster, params);
    },

    openBulkConfigureCluster: (clusterNames: string[]) => {
      setPanelParams(panels.bulkConfigureCluster, {
        clusters: clusterNames.join(","),
      });
    },
  };
};

export default usePanelParams;
