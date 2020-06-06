import React from "react";
import { useRecoilValue } from "recoil";
import styled from "./styles/styled";
import { Page, page as pageAtom } from "./atoms/page";
import DownloadsPage from "./pages/Downloads";
import SearchPage from "./pages/Search";

const Container = styled.main`
  flex-grow: 1;
  height: auto;
  max-width: 100%;
  background-color: ${(props) => props.theme.color.workspaceBackground};
  overflow: hidden;
`;

const Workspace = () => {
  const page = useRecoilValue(pageAtom);

  return (
    <Container>
      {page === Page.Downloads && <DownloadsPage />}
      {page === Page.Search && <SearchPage />}
    </Container>
  );
};

export default React.memo(Workspace);
