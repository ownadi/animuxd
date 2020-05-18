import React, { useState, useCallback } from "react";
import styled from "./styles/styled";
import { Global, css } from "@emotion/core";
import { ThemeProvider } from "emotion-theming";
import theme from "./styles/theme";
import { FaSearch, FaDownload } from "react-icons/fa";
import DownloadsPage from "./pages/Downloads";
import SearchPage from "./pages/Search";

const PAGE_DOWNLOADS = "PAGE_DOWNLOADS";
const PAGE_SEARCH = "PAGE_SEARCH";

const Window = styled.div`
  width: 100vw;
  height: 100vh;
  display: flex;

  > * {
    padding: 1rem;
    box-sizing: border-box;
  }
`;

const Navigation = styled.nav`
  height: 100%;
  width: 7rem;
  background-color: ${(props) => props.theme.color.navigationBackground};
  color: ${(props) => props.theme.color.navigationTextColor};
  text-align: center;
`;

const NavigationList = styled.ul`
  list-style: none;
  margin: 0;
  padding: 0;
  font-size: 2.5rem;
`;

type NavigationListItemProps = {
  active: boolean;
};
const NavigationListItem = styled.li<NavigationListItemProps>`
  display: flex;
  justify-content: center;
  align-items: center;
  background-color: ${(props) => props.theme.color.navigationItemBackground};
  margin-bottom: 2rem;
  height: 5rem;
  border-radius: 50%;
  cursor: pointer;
  transition: all 0.2s ease-in-out;

  ${(props) =>
    !props.active &&
    css`
      &:hover {
        filter: brightness(125%);
      }
    `}

  ${(props) =>
    props.active &&
    css`
      border-radius: 1.6rem;
      background-color: ${props.theme.color.navigationActiveItemBackground};
      color: ${props.theme.color.navigationActiveColor};
    `}
`;

const Workspace = styled.main`
  flex-grow: 1;
  height: 100%;
  background-color: ${(props) => props.theme.color.workspaceBackground};
  overflow: hidden;
`;

function App() {
  const [page, setPage] = useState(PAGE_SEARCH);

  const navToDownloads = useCallback(() => setPage(PAGE_DOWNLOADS), [setPage]);
  const navToSearch = useCallback(() => setPage(PAGE_SEARCH), [setPage]);

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
        <Navigation>
          <NavigationList>
            <NavigationListItem
              onClick={navToDownloads}
              active={page === PAGE_DOWNLOADS}
            >
              <FaDownload />
            </NavigationListItem>
            <NavigationListItem
              onClick={navToSearch}
              active={page === PAGE_SEARCH}
            >
              <FaSearch />
            </NavigationListItem>
          </NavigationList>
        </Navigation>
        <Workspace>
          {page === PAGE_DOWNLOADS && <DownloadsPage />}
          {page === PAGE_SEARCH && <SearchPage />}
        </Workspace>
      </Window>
    </ThemeProvider>
  );
}

export default App;
