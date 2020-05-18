import styled from "./styled";

export const Container = styled.div`
  overflow-y: scroll;
  flex-grow: 1;

  &::-webkit-scrollbar {
    width: 0.5rem;
    height: 0.5rem;
  }

  &::-webkit-scrollbar-track {
    background-color: ${(props) => props.theme.color.scrollTrack};
  }

  &::-webkit-scrollbar-thumb {
    background-color: ${(props) => props.theme.color.scrollThumb};
    border-radius: 1.6rem;
  }

  &::-webkit-scrollbar-button {
    display: none;
  }
`;

export const Table = styled.table`
  width: 100%;
  color: ${(props) => props.theme.color.tableText};
`;

export const Header = styled.th`
  padding: 1rem;
  text-align: left;
  position: sticky;
  top: 0;
  box-sizing: border-box;
  background-color: ${(props) => props.theme.color.tableHeader};
  user-select: none;
`;

export const Data = styled.td`
  padding: 1rem;
  box-sizing: border-box;
`;

export const Row = styled.tr`
  background-color: ${(props) => props.theme.color.tableRow};

  &:hover {
    background-color: ${(props) => props.theme.color.tableRowHover};
  }
`;
