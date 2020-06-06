import React, { useCallback } from "react";
import styled from "./styles/styled";
import css from "@emotion/css";
import { FaSearch, FaDownload } from "react-icons/fa";
import { useRecoilState } from "recoil";
import { Page, page as pageAtom } from "./atoms/page";

const Container = styled.nav`
  height: 100%;
  width: 7rem;
  min-width: 7rem;
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

const Navigation = () => {
  const [page, setPage] = useRecoilState<Page>(pageAtom);

  const navToDownloads = useCallback(() => setPage(Page.Downloads), [setPage]);
  const navToSearch = useCallback(() => setPage(Page.Search), [setPage]);

  return (
    <Container>
      <NavigationList>
        <NavigationListItem
          onClick={navToDownloads}
          active={page === Page.Downloads}
        >
          <FaDownload />
        </NavigationListItem>
        <NavigationListItem onClick={navToSearch} active={page === Page.Search}>
          <FaSearch />
        </NavigationListItem>
      </NavigationList>
    </Container>
  );
};

export default React.memo(Navigation);
