import React from "react";
import styled from "./styles/styled";
import { Global, css } from "@emotion/core";
import { ThemeProvider } from "emotion-theming";
import theme from "./styles/theme";
import { RecoilRoot } from "recoil";
import useDownloads from "./hooks/useDownloads";
import Navigation from "./Navigation";
import Workspace from "./Workspace";
import Status from "./Status";

const Window = styled.div`
  width: 100vw;
  height: 100vh;
  display: flex;
  flex-direction: column;
`;

const Content = styled.div`
  display: flex;
  flex-grow: 1;
  max-height: calc(100vh - 2.5rem);

  > * {
    padding: 1rem;
    box-sizing: border-box;
  }
`;

function App() {
  useDownloads();

  return (
    <ThemeProvider theme={theme}>
      <Global
        styles={css`
          html {
            @import url("https://fonts.googleapis.com/css2?family=Roboto&display=swap");
            font-family: "Roboto", sans-serif;
            font-size: 62.5%;
          }

          body {
            margin: 0;
            font-size: 1.6rem;
          }

          .fa-spin {
            animation: fa-spin 2s infinite linear;
          }

          @keyframes fa-spin {
            0% {
              transform: rotate(0deg);
            }
            100% {
              transform: rotate(359deg);
            }
          }
        `}
      />
      <Window>
        <Content>
          <Navigation />
          <Workspace />
        </Content>
        <Status />
      </Window>
    </ThemeProvider>
  );
}

export default () => (
  <RecoilRoot>
    <App />
  </RecoilRoot>
);
