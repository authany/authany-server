import React, { useCallback, useContext, useMemo } from "react";
import { useParams } from "react-router-dom";
import { Context } from "./intl";
import {
  Text,
  CommandButton,
  IconButton,
  Panel,
  PanelType,
  IRenderFunction,
  IPanelProps,
  IContextualMenuProps,
} from "@fluentui/react";
import { useViewerQuery } from "./graphql/portal/query/viewerQuery";
import ScreenNav from "./ScreenNav";
import Link from "./Link";

import styles from "./ScreenHeader.module.css";
import { useSystemConfig } from "./context/SystemConfigContext";
import { useBoolean } from "@fluentui/react-hooks";
import ExternalLink from "./ExternalLink";
import { useLogout } from "./graphql/portal/Authenticated";
import { useSettingsAnchor } from "./hook/authgear";
import { Logo } from "./components/common/Logo";
import logoStyles from "./components/common/Logo.module.css";
import ProjectSelector from "./components/header/ProjectSelector";
import { LOCALE_DISPLAY_NAMES, useLocale } from "./locale";

interface HeaderAppSectionProps {
  appID: string;
}

const HeaderAppSection: React.VFC<HeaderAppSectionProps> = (props) => {
  const { appID } = props;
  const { themes } = useSystemConfig();

  return (
    <>
      <span
        className={styles.headerDivider}
        role="separator"
        aria-hidden={true}
      />
      <ProjectSelector appID={appID} theme={themes.inverted} />
    </>
  );
};

const commandButtonStyles = {
  label: {
    fontSize: "12px",
  },
  menuIcon: {
    fontSize: "12px",
    color: "white",
  },
};

interface MobileViewHeaderIconSectionProps {
  onClick: () => void;
  showHamburger: boolean;
}

const MobileViewHeaderIconSection: React.VFC<
  MobileViewHeaderIconSectionProps
> = (props) => {
  const { onClick, showHamburger } = props;
  const { themes } = useSystemConfig();

  return (
    <>
      {showHamburger ? (
        <IconButton
          ariaLabel="hamburger"
          iconProps={{ iconName: "WaffleOffice365" }}
          className={styles.hamburger}
          theme={themes.inverted}
          onClick={onClick}
        />
      ) : (
        <Link to="/" className={styles.logoLink}>
          <Logo containerClassName={logoStyles.logo__containerHeader} />
        </Link>
      )}
    </>
  );
};

const DesktopViewHeaderIconSection: React.VFC = () => {
  return (
    <Link to="/" className={styles.logoLink}>
      <Logo containerClassName={logoStyles.logo__containerHeader} />
    </Link>
  );
};

const MobileViewNavbarHeader: IRenderFunction<IPanelProps> = (props) => {
  // eslint-disable-next-line @typescript-eslint/no-non-null-asserted-optional-chain
  const onClick: () => void = props?.onDismiss!;
  return (
    <div className={styles.headerMobile}>
      <IconButton
        ariaLabel="hamburger"
        iconProps={{ iconName: "WaffleOffice365" }}
        className={styles.hamburger}
        onClick={onClick}
      />
      <Logo inverted={true} />
    </div>
  );
};

const MobileViewNavbarBody: IRenderFunction<IPanelProps> = (props) => {
  // eslint-disable-next-line @typescript-eslint/no-non-null-asserted-optional-chain
  const onClick: () => void = props?.onDismiss!;
  return <ScreenNav mobileView={true} onLinkClick={onClick} />;
};

interface ScreenNavProps {
  showHamburger?: boolean;
}

const ScreenHeader: React.VFC<ScreenNavProps> = function ScreenHeader(props) {
  const { showHamburger = true } = props;
  const { renderToString } = useContext(Context);
  const { themes } = useSystemConfig();
  const { appID } = useParams() as { appID: string };
  const { viewer } = useViewerQuery();
  const [isNavbarOpen, { setTrue: openNavbar, setFalse: dismissNavbar }] =
    useBoolean(false);

  const logout = useLogout();

  const onClickLogout = useCallback(() => {
    logout().catch((err: unknown) => {
      console.error(err);
    });
  }, [logout]);

  const onClickCookiePreference = useCallback(() => {
    if (window.Osano?.cm !== undefined) {
      window.Osano.cm.showDrawer("osano-cm-dom-info-dialog-open");
    } else {
      console.error("Osano is not loaded");
    }
  }, []);

  const headerStyle = useMemo(
    () => ({
      backgroundColor: themes.main.palette.themePrimary,
    }),
    [themes.main]
  );

  const { href: settingURL, onClick: onClickSettings } = useSettingsAnchor();
  const { locale, availableLocales, setLocale } = useLocale();

  const menuProps = useMemo(() => {
    const items = [
      {
        key: "settings",
        text: renderToString("ScreenHeader.settings"),
        iconProps: {
          iconName: "PlayerSettings",
        },
        href: settingURL,
        onClick: onClickSettings,
      },
      {
        key: "language",
        text: renderToString("ScreenHeader.language"),
        iconProps: {
          iconName: "LocaleLanguage",
        },
        subMenuProps: {
          items: availableLocales.map((l) => ({
            key: `language-${l}`,
            text: LOCALE_DISPLAY_NAMES[l] ?? l,
            canCheck: true,
            checked: l === locale,
            onClick: () => {
              setLocale(l);
            },
          })),
        },
      },
      {
        key: "logout",
        text: renderToString("ScreenHeader.sign-out"),
        iconProps: {
          iconName: "SignOut",
        },
        onClick: onClickLogout,
      },
    ] satisfies IContextualMenuProps["items"];

    if (window.Osano !== undefined) {
      items.splice(1, 0, {
        key: "cookie",
        text: renderToString("ScreenHeader.cookie-preference"),
        iconProps: {
          iconName: "Cookies",
        },
        onClick: onClickCookiePreference,
      });
    }

    return { items } satisfies IContextualMenuProps;
  }, [
    renderToString,
    settingURL,
    onClickSettings,
    onClickLogout,
    onClickCookiePreference,
    locale,
    availableLocales,
    setLocale,
  ]);

  return (
    <header className={styles.header} style={headerStyle}>
      <div className={styles.mobileView}>
        <MobileViewHeaderIconSection
          showHamburger={showHamburger}
          onClick={openNavbar}
        />
        {appID ? <HeaderAppSection appID={appID} /> : null}
        <Panel
          isLightDismiss={true}
          hasCloseButton={false}
          isOpen={isNavbarOpen}
          onDismiss={dismissNavbar}
          type={PanelType.smallFixedNear}
          onRenderNavigation={MobileViewNavbarHeader}
          onRenderBody={MobileViewNavbarBody}
        />
      </div>
      <div className={styles.desktopView}>
        <DesktopViewHeaderIconSection />
        {appID ? <HeaderAppSection appID={appID} /> : null}
      </div>
      <div className={styles.links}>
        {/* Authany: header docs link points at our own documentation site */}
        <ExternalLink href="https://docs.authany.com/" className={styles.link}>
          <Text variant="small">
            {renderToString("ScreenHeader.links.documentation")}
          </Text>
        </ExternalLink>
      </div>
      {viewer != null ? (
        <CommandButton
          className={styles.desktopView}
          menuProps={menuProps}
          theme={themes.inverted}
          styles={commandButtonStyles}
        >
          {viewer.email}
        </CommandButton>
      ) : null}
    </header>
  );
};

export default ScreenHeader;
