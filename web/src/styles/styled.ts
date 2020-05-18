import styled, { CreateStyled } from "@emotion/styled/macro";

export type Theme = {
  color: {
    navigationBackground: string;
    navigationItemBackground: string;
    navigationActiveItemBackground: string;
    navigationTextColor: string;
    navigationActiveColor: string;
    workspaceBackground: string;
    inputBorder: string;
    inputFocusBorder: string;
    inputBackground: string;
    inputText: string;
    tableText: string;
    tableHeader: string;
    tableRow: string;
    tableRowHover: string;
    scrollTrack: string;
    scrollThumb: string;
  };
};

export default styled as CreateStyled<Theme>;
