import { act } from "react";
import { createRoot } from "react-dom/client";
import { MemoryRouter, useLocation } from "react-router-dom";
import ClusterSearchFilter from "./ClusterSearchFilter";

describe("ClusterSearchFilter", () => {
  beforeAll(() => {
    (
      globalThis as unknown as { IS_REACT_ACT_ENVIRONMENT: boolean }
    ).IS_REACT_ACT_ENVIRONMENT = true;
  });

  it("keeps the filters from the URL when it mounts", () => {
    let search = "";
    const LocationSpy = () => {
      search = useLocation().search;
      return null;
    };

    const container = document.createElement("div");
    const root = createRoot(container);

    act(() => {
      root.render(
        <MemoryRouter
          initialEntries={["/ui/clusters?query=prod&node-status=Online"]}
        >
          <ClusterSearchFilter />
          <LocationSpy />
        </MemoryRouter>,
      );
    });

    expect(search).toBe("?query=prod&node-status=Online");

    act(() => {
      root.unmount();
    });
  });
});
