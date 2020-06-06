import React from "react";
import filesize from "filesize";
import styled from "./styles/styled";
import { useRecoilValue } from "recoil";
import { sumCurrentSpeed, waiting, downloading } from "./atoms/downloads";

const Container = styled.div`
  height: 2.5rem;
  font-size: 0.8em;
  background-color: ${(props) => props.theme.color.statusBackground};
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 1rem;
  color: ${(props) => props.theme.color.statusText};
  border-top: 0.1rem solid ${(props) => props.theme.color.navigationBackground};

  > div {
    display: flex;
  }
`;

const Group = styled.div`
  &:not(:last-of-type) {
    margin-right: 2rem;
  }
`;

const Status = () => {
  const inProgressDownloads = useRecoilValue(downloading);
  const waitingDownloads = useRecoilValue(waiting);
  const summaricCurrentSpeed = useRecoilValue(sumCurrentSpeed);

  return (
    <Container>
      <div />
      <div>
        <Group>
          Downloading: {inProgressDownloads.length} | Waiting:{" "}
          {waitingDownloads.length}
        </Group>
        <Group style={{ minWidth: "7.5rem", textAlign: "right" }}>
          {filesize(summaricCurrentSpeed) + "/s"}
        </Group>
      </div>
    </Container>
  );
};

export default Status;
