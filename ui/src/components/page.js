import React from 'react';
import { NotFoundBoundary, useCurrentRoute } from 'react-navi';
import { createGlobalStyle } from 'styled-components';
import { Helmet } from 'react-helmet-async';

import { Box } from './core';
import NotFound from './not-found';

const GlobalStyle = createGlobalStyle`
  * {
    box-sizing: inherit;
  }

  ::selection {
    background-color: rgba(255,255,255,.99);
    color: #000;
  }

  html {
    height: 100%;
    box-sizing: border-box;
    font-family: ${props => props.theme.fonts.default};
    font-size: 16px;
    font-weight: 400;
    line-height: 1.2;
    background-color: ${props => props.theme.colors.black};
  }

  html, body {
    text-rendering: optimizelegibility;
    shape-rendering: geometricprecision;
    -webkit-font-smoothing: antialiased;
  }

  body {
    height: 100%;
    overflow: hidden !important;
    margin: 0;
  }

  main {
    height: 100%;
  }

  strong, strong * {
    font-weight: 500;
  }

  .ace_editor {
    font-family: ${props => props.theme.fonts.code} !important;
    background: ${props => props.theme.colors.grays[0]} !important;
    color: ${props => props.theme.colors.white} !important;
  }
  .ace_gutter {
    background: ${props => props.theme.colors.grays[1]} !important;
    color: ${props => props.theme.colors.white} !important;
  }
  .ace_gutter-active-line {
    background: ${props => props.theme.colors.grays[2]} !important;
  }
  .ace_active-line {
    background: ${props => props.theme.colors.grays[2]} !important;
  }
  .ace_cursor {
    border-left: 1px solid ${props => props.theme.colors.primary} !important;
  }
  .ace_gutter-cell {
    padding-left: 0 !important;
  }

  svg[data-icon="caret-down"] {
    fill: ${props => props.theme.colors.white} !important;
  }
`;

const Page = ({ children }) => {
  const route = useCurrentRoute();
  return (
    <>
      <Helmet>
        {route.title && <title>{`${route.title} - Deviceplane`}</title>}
        <link
          href="https://fonts.googleapis.com/css?family=Rubik:300,400,500,700&display=swap"
          rel="stylesheet"
        />
      </Helmet>
      <GlobalStyle />
      <NotFoundBoundary
        render={() => {
          if (route.data.context.currentUser) {
            return <NotFound />;
          } else {
            window.location.replace('/login');
            return null;
          }
        }}
      >
        {children}
      </NotFoundBoundary>
    </>
  );
};

export default Page;
