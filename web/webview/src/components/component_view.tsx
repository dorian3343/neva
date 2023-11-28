import {
  VSCodeDataGrid,
  VSCodeDataGridRow,
  VSCodeDataGridCell,
} from "@vscode/webview-ui-toolkit/react";
import { EntityRef } from "../generated/sourcecode";
import { InterfaceView } from "./interface_view";
import NetView from "./network_view";
import { ComponentViewState, NodesViewState } from "../core/file_view_state";

export function ComponentView(props: {
  name: string;
  entity: ComponentViewState;
  style?: object;
}) {
  return (
    <div style={props.style}>
      <h3
        style={{ marginBottom: "10px", display: "flex", alignItems: "center" }}
      >
        {props.name}
      </h3>
      {props.entity.interface && (
        <InterfaceView name="" entity={props.entity.interface} />
      )}
      {props.entity.nodes.length > 0 && (
        <NodesView nodes={props.entity.nodes} />
      )}
      {props.entity.nodes.length > 0 &&
        props.entity.interface &&
        props.entity.net && <NetView component={props.entity} />}
    </div>
  );
}

function NodesView(props: { nodes: NodesViewState[] }) {
  // TODO render their ports
  return (
    <>
      <h4>Nodes</h4>
      <VSCodeDataGrid>
        {props.nodes.map(({ name, node }) => (
          <VSCodeDataGridRow>
            <VSCodeDataGridCell grid-column="1">{name}</VSCodeDataGridCell>
            <VSCodeDataGridCell grid-column="2">
              {formatEntityRef(node.entityRef)}
            </VSCodeDataGridCell>
          </VSCodeDataGridRow>
        ))}
      </VSCodeDataGrid>
    </>
  );
}

function formatEntityRef(ref?: EntityRef): string {
  if (!ref) {
    return "";
  }
  if (!ref.pkg) {
    return String(ref.name);
  }
  return String(ref.pkg) + "." + String(ref.name);
}
