import React, {
  useMemo,
  useContext,
  useState,
  useCallback,
  useRef,
} from "react";
import cn from "classnames";
import {
  Avatar,
  Badge,
  DropdownMenu,
  IconButton as RadixIconButton,
  Text,
} from "@radix-ui/themes";
import {
  CaretDownIcon,
  CaretSortIcon,
  CaretUpIcon,
  DotsVerticalIcon,
  EyeNoneIcon,
  LockClosedIcon,
  Pencil1Icon,
  TrashIcon,
} from "@radix-ui/react-icons";
import { Context, FormattedMessage } from "../../intl";
import { useNavigate, useParams } from "react-router-dom";
import { UsersListFragment } from "./query/usersListQuery.generated";
import {
  UserSortBy,
  SortDirection,
  Role,
  Group,
} from "./globalTypes.generated";

import PaginationWidget from "../../PaginationWidget";
import {
  AccountStatusDialog,
  AccountStatusDialogProps,
} from "./UserDetailsAccountStatus";

import { extractRawID } from "../../util/graphql";
import { formatDatetime } from "../../util/formatDatetime";
import { formatDateOnly } from "../../util/formatDateOnly";
import { Tooltip } from "../../components/v2/Tooltip/Tooltip";
import { CopyIconButton } from "../../components/v2/CopyIconButton/CopyIconButton";

import styles from "./UsersList.module.css";
import { useDebounced } from "../../hook/useDebounced";
import { useElementWidth } from "../../hook/useElementWidth";

interface UsersListProps {
  className?: string;
  isSearch: boolean;
  loading: boolean;
  users: UsersListFragment | null;
  offset: number;
  pageSize: number;
  totalCount?: number;
  onChangeOffset?: (offset: number) => void;
  onColumnClick?: (columnKey: UserSortBy) => void;
  sortBy?: UserSortBy;
  sortDirection?: SortDirection;
  showRolesAndGroups: boolean;
}

interface UserListRoleItem extends Pick<Role, "id" | "name" | "key"> {}
interface UserListGroupItem extends Pick<Group, "id" | "name" | "key"> {}

interface UserListRoles {
  totalCount: number;
  items: UserListRoleItem[];
}

interface UserListGroups {
  totalCount: number;
  items: UserListGroupItem[];
}

interface UserListItem {
  id: string;
  rawID: string;
  isAnonymous: boolean;
  isAnonymized: boolean;
  isDisabled: boolean;
  isDeactivated: boolean;
  deleteAt: string | null;
  anonymizeAt: string | null;
  accountValidFrom: string | null;
  accountValidUntil: string | null;
  temporarilyDisabledFrom: string | null;
  temporarilyDisabledUntil: string | null;
  createdAt: string | null;
  createdAtDateOnly: string | null;
  lastLoginAt: string | null;
  lastLoginAtDateOnly: string | null;
  profilePictureURL: string | null;
  formattedName: string | null;
  endUserAccountID: string | null;
  username: string | null;
  phone: string | null;
  email: string | null;
  roles: UserListRoles;
  groups: UserListGroups;
}

interface DisableUserDialogData {
  accountStatus: UserListItem;
  mode: AccountStatusDialogProps["mode"];
}

const USER_LIST_PLACEHOLDER = "-";

// List widths (px) below which columns fold into the user cell: the sum of
// the column minimums in UsersList.module.css at each stage.
const FOLD_WIDTHS = [1056, 800, 656];
const FOLD_WIDTHS_WITH_ROLES = [1296, 1040, 896];

function foldLevel(width: number, showRolesAndGroups: boolean): number {
  if (width === 0) {
    return 0;
  }
  const thresholds = showRolesAndGroups ? FOLD_WIDTHS_WITH_ROLES : FOLD_WIDTHS;
  return thresholds.filter((threshold) => width < threshold).length;
}

interface UserInfoProps {
  item: UserListItem;
  showRolesAndGroups: boolean;
}

function stopPropagation(e: React.SyntheticEvent) {
  e.stopPropagation();
}

function UserInfo(props: UserInfoProps) {
  const {
    item: {
      profilePictureURL,
      formattedName,
      endUserAccountID,
      rawID,
      isAnonymous,
      isAnonymized,
      isDisabled,
      username,
      phone,
      email,
      roles,
      groups,
    },
    showRolesAndGroups,
  } = props;
  const nameText = formattedName ?? endUserAccountID ?? rawID;
  const displayName = isAnonymous ? (
    <Text className={styles.anonymousUserLabel}>
      <FormattedMessage id="UsersList.anonymous-user" />
    </Text>
  ) : isAnonymized ? (
    <Text className={styles.anonymizedUserLabel}>
      <FormattedMessage id="UsersList.anonymized-user" />
    </Text>
  ) : (
    nameText
  );
  const fallback = nameText.trim().charAt(0).toUpperCase() || "?";
  const relatedItemsText = showRolesAndGroups
    ? [roles, groups]
        .filter((related) => related.totalCount > 0)
        .map(getRelatedItemsText)
        .join(" · ")
    : "";

  return (
    <div className={styles.userInfo}>
      <Avatar
        className={styles.userInfoPicture}
        size="3"
        radius="full"
        src={profilePictureURL ?? undefined}
        fallback={fallback}
      />
      <div className={styles.userInfoLines}>
        <div className={styles.userInfoDisplayName}>
          <Text
            size="2"
            weight="medium"
            className={styles.userInfoDisplayNameText}
          >
            {displayName}
          </Text>
          {isDisabled && !isAnonymized ? (
            <Badge
              color="red"
              size="1"
              radius="small"
              className={styles.disabledBadge}
            >
              <FormattedMessage id="AccountStatusBadge.disabled" />
            </Badge>
          ) : null}
        </div>
        {/* Values of columns folded away on narrow lists; hidden otherwise. */}
        <Text size="1" color="gray" className={styles.userInfoMeta}>
          {email != null && email !== nameText ? (
            <span className={cn(styles.metaItem, styles["metaItem--email"])}>
              {email}
            </span>
          ) : null}
          {username != null ? (
            <span className={cn(styles.metaItem, styles["metaItem--username"])}>
              {username}
            </span>
          ) : null}
          {phone != null ? (
            <span className={cn(styles.metaItem, styles["metaItem--phone"])}>
              {phone}
            </span>
          ) : null}
          {relatedItemsText !== "" ? (
            <span className={cn(styles.metaItem, styles["metaItem--roles"])}>
              {relatedItemsText}
            </span>
          ) : null}
        </Text>
        <div className={styles.userInfoID}>
          <Tooltip content={rawID}>
            <Text size="1" color="gray" className={styles.userInfoRawID}>
              {rawID}
            </Text>
          </Tooltip>
          <span
            className={styles.copyID}
            onClick={stopPropagation}
            onKeyDown={stopPropagation}
          >
            <CopyIconButton textToCopy={rawID} />
          </span>
        </div>
      </div>
    </div>
  );
}

function CellText({ value }: { value: string | null }) {
  return (
    <Text size="2" className={styles.cellText}>
      {value ?? USER_LIST_PLACEHOLDER}
    </Text>
  );
}

function UserTextCell({
  value,
  column,
}: {
  value: string | null;
  column: string;
}) {
  return (
    <div className={cn(styles.cell, styles[`col--${column}`])}>
      <CellText value={value} />
    </div>
  );
}

// Shows the date only, with the full datetime (incl. timezone) on hover.
function DateCell({
  dateOnly,
  datetime,
  column,
}: {
  dateOnly: string | null;
  datetime: string | null;
  column: string;
}) {
  const className = cn(
    styles.cell,
    styles["cell--date"],
    styles[`col--${column}`]
  );
  if (dateOnly == null || datetime == null) {
    return (
      <div className={className}>
        <CellText value={dateOnly} />
      </div>
    );
  }
  return (
    <div className={className}>
      <Tooltip content={datetime}>
        <Text size="2" className={styles.cellText}>
          {dateOnly}
        </Text>
      </Tooltip>
    </div>
  );
}

function getRelatedItemsText(
  relatedItems: UserListRoles | UserListGroups
): string {
  if (relatedItems.totalCount === 0) {
    return USER_LIST_PLACEHOLDER;
  }
  const firstItem = relatedItems.items[0];
  const additionalCount = relatedItems.totalCount - 1;
  return `${firstItem.name ?? firstItem.key}${
    additionalCount > 0 ? ` +${additionalCount}` : ""
  }`;
}

function SortHeader({
  active,
  sortDirection,
  onClick,
  children,
}: {
  active: boolean;
  sortDirection?: SortDirection;
  onClick: () => void;
  children: React.ReactNode;
}) {
  const SortIcon = !active
    ? CaretSortIcon
    : sortDirection === SortDirection.Asc
    ? CaretUpIcon
    : CaretDownIcon;
  return (
    <button
      type="button"
      className={cn(styles.sortButton, active && styles["sortButton--active"])}
      onClick={onClick}
    >
      {children}
      <SortIcon className={styles.sortIcon} />
    </button>
  );
}

const UsersList: React.VFC<UsersListProps> = function UsersList(props) {
  const {
    className,
    isSearch,
    loading: rawLoading,
    offset,
    pageSize,
    totalCount,
    onChangeOffset,
    onColumnClick,
    sortBy,
    sortDirection,
    showRolesAndGroups,
  } = props;
  const edges = props.users?.edges;

  const [loading] = useDebounced(rawLoading, 500);
  const rootRef = useRef<HTMLDivElement>(null);
  const rootWidth = useElementWidth(rootRef);

  const { renderToString, locale } = useContext(Context);
  const { appID } = useParams() as { appID: string };
  const navigate = useNavigate();

  const [isDisableUserDialogHidden, setIsDisableUserDialogHidden] =
    useState(true);
  const [disableUserDialogData, setDisableUserDialogData] =
    useState<DisableUserDialogData | null>(null);

  const items: UserListItem[] = useMemo(() => {
    const items: UserListItem[] = [];
    if (edges != null) {
      for (const edge of edges) {
        const node = edge?.node;
        if (node != null) {
          items.push({
            id: node.id,
            rawID: extractRawID(node.id),
            isAnonymous: node.isAnonymous,
            isAnonymized: node.isAnonymized,
            isDisabled: node.isDisabled,
            isDeactivated: node.isDeactivated,
            deleteAt: formatDatetime(locale, node.deleteAt),
            anonymizeAt: formatDatetime(locale, node.anonymizeAt),
            accountValidFrom: node.accountValidFrom,
            accountValidUntil: node.accountValidUntil,
            temporarilyDisabledFrom: node.temporarilyDisabledFrom,
            temporarilyDisabledUntil: node.temporarilyDisabledUntil,
            createdAt: formatDatetime(locale, node.createdAt),
            createdAtDateOnly: formatDateOnly(locale, node.createdAt),
            lastLoginAt: formatDatetime(locale, node.lastLoginAt),
            lastLoginAtDateOnly: formatDateOnly(locale, node.lastLoginAt),
            profilePictureURL: node.standardAttributes.picture ?? null,
            formattedName: node.formattedName ?? null,
            endUserAccountID: node.endUserAccountID ?? null,
            username: node.standardAttributes.preferred_username ?? null,
            phone: node.standardAttributes.phone_number ?? null,
            email: node.standardAttributes.email ?? null,
            roles: {
              totalCount: node.effectiveRoles?.totalCount ?? 0,
              items: (node.effectiveRoles?.edges ?? []).flatMap(
                (edge) => edge?.node ?? []
              ),
            },
            groups: {
              totalCount: node.groups?.totalCount ?? 0,
              items: (node.groups?.edges ?? []).flatMap(
                (edge) => edge?.node ?? []
              ),
            },
          });
        }
      }
    }
    return items;
  }, [edges, locale]);

  const openAccountStatusDialog = useCallback(
    (item: UserListItem, mode: AccountStatusDialogProps["mode"]) => {
      setDisableUserDialogData({
        accountStatus: item,
        mode,
      });
      setIsDisableUserDialogHidden(false);
    },
    []
  );

  const renderActionCell = useCallback(
    (item: UserListItem) => {
      return (
        <DropdownMenu.Root>
          <DropdownMenu.Trigger>
            <RadixIconButton
              className={styles.rowActionsButton}
              variant="soft"
              color="gray"
              size="2"
              aria-label={renderToString("action")}
            >
              <DotsVerticalIcon width="1rem" height="1rem" />
            </RadixIconButton>
          </DropdownMenu.Trigger>
          <DropdownMenu.Content align="end">
            <DropdownMenu.Item
              onSelect={() => {
                navigate(`/project/${appID}/users/${item.id}/details`);
              }}
            >
              <Pencil1Icon />
              <FormattedMessage id="edit" />
            </DropdownMenu.Item>
            <DropdownMenu.Separator />
            <DropdownMenu.Item
              color="red"
              onSelect={() => {
                openAccountStatusDialog(item, "disable");
              }}
            >
              <LockClosedIcon />
              <FormattedMessage id="disable" />
            </DropdownMenu.Item>
            <DropdownMenu.Item
              color="red"
              onSelect={() => {
                openAccountStatusDialog(item, "anonymize-or-schedule");
              }}
            >
              <EyeNoneIcon />
              <FormattedMessage id="UserDetailsAccountStatus.anonymize-user.action.anonymize" />
            </DropdownMenu.Item>
            <DropdownMenu.Item
              color="red"
              onSelect={() => {
                openAccountStatusDialog(item, "delete-or-schedule");
              }}
            >
              <TrashIcon />
              <FormattedMessage id="delete" />
            </DropdownMenu.Item>
          </DropdownMenu.Content>
        </DropdownMenu.Root>
      );
    },
    [appID, navigate, openAccountStatusDialog, renderToString]
  );

  const dismissDisableUserDialog = useCallback(() => {
    setIsDisableUserDialogHidden(true);
  }, []);

  const isEmpty = !loading && items.length === 0;
  const fold = foldLevel(rootWidth, showRolesAndGroups);
  const hasUsername = items.some((item) => item.username != null);
  const hasPhone = items.some((item) => item.phone != null);
  const headerCell = (column: string, extra?: string) =>
    cn(styles.headerCell, styles[`col--${column}`], extra);

  return (
    <>
      <div ref={rootRef} className={cn(styles.root, className)}>
        <div className={styles.tableWrapper}>
          <div
            className={cn(
              styles.table,
              showRolesAndGroups && styles["table--withRoles"],
              !loading && !hasUsername && styles["table--noUsername"],
              !loading && !hasPhone && styles["table--noPhone"],
              fold > 0 && styles[`table--fold${fold}`]
            )}
          >
            <div className={styles.tableHeader}>
              <div className={headerCell("user", styles["headerCell--user"])}>
                <FormattedMessage id="UsersList.column.raw-id" />
              </div>
              <div className={headerCell("username")}>
                <FormattedMessage id="UsersList.column.username" />
              </div>
              <div className={headerCell("email")}>
                <FormattedMessage id="UsersList.column.email" />
              </div>
              <div className={headerCell("phone")}>
                <FormattedMessage id="UsersList.column.phone" />
              </div>
              {showRolesAndGroups ? (
                <>
                  <div className={headerCell("roles")}>
                    <FormattedMessage id="UsersList.column.roles" />
                  </div>
                  <div className={headerCell("groups")}>
                    <FormattedMessage id="UsersList.column.groups" />
                  </div>
                </>
              ) : null}
              <div
                className={headerCell("created", styles["headerCell--sort"])}
              >
                <SortHeader
                  active={sortBy === UserSortBy.CreatedAt}
                  sortDirection={sortDirection}
                  onClick={() => onColumnClick?.(UserSortBy.CreatedAt)}
                >
                  <FormattedMessage id="UsersList.column.signed-up" />
                </SortHeader>
              </div>
              <div
                className={headerCell("last-login", styles["headerCell--sort"])}
              >
                <SortHeader
                  active={sortBy === UserSortBy.LastLoginAt}
                  sortDirection={sortDirection}
                  onClick={() => onColumnClick?.(UserSortBy.LastLoginAt)}
                >
                  <FormattedMessage id="UsersList.column.last-login-at" />
                </SortHeader>
              </div>
              <div
                className={headerCell("action", styles["headerCell--action"])}
              />
            </div>
            {loading
              ? Array.from({ length: 5 }).map((_, index) => (
                  <div key={index} className={styles.shimmerRow}>
                    <div className={styles.shimmerBlock} />
                  </div>
                ))
              : items.map((item) => (
                  <div
                    key={item.id}
                    className={styles.tableRow}
                    role="button"
                    tabIndex={0}
                    onClick={() => {
                      navigate(`/project/${appID}/users/${item.id}/details`);
                    }}
                    onKeyDown={(e) => {
                      if (e.key === "Enter" || e.key === " ") {
                        e.preventDefault();
                        navigate(`/project/${appID}/users/${item.id}/details`);
                      }
                    }}
                  >
                    <div
                      className={cn(
                        styles.cell,
                        styles["cell--user"],
                        styles["col--user"]
                      )}
                    >
                      <UserInfo
                        item={item}
                        showRolesAndGroups={showRolesAndGroups}
                      />
                    </div>
                    <UserTextCell value={item.username} column="username" />
                    <UserTextCell value={item.email} column="email" />
                    <UserTextCell value={item.phone} column="phone" />
                    {showRolesAndGroups ? (
                      <>
                        <UserTextCell
                          value={getRelatedItemsText(item.roles)}
                          column="roles"
                        />
                        <UserTextCell
                          value={getRelatedItemsText(item.groups)}
                          column="groups"
                        />
                      </>
                    ) : null}
                    <DateCell
                      dateOnly={item.createdAtDateOnly}
                      datetime={item.createdAt}
                      column="created"
                    />
                    <DateCell
                      dateOnly={item.lastLoginAtDateOnly}
                      datetime={item.lastLoginAt}
                      column="last-login"
                    />
                    <div
                      className={cn(
                        styles.cell,
                        styles["cell--action"],
                        styles["col--action"]
                      )}
                      onClick={(e) => e.stopPropagation()}
                      onKeyDown={(e) => e.stopPropagation()}
                    >
                      {renderActionCell(item)}
                    </div>
                  </div>
                ))}
            {isEmpty ? (
              <div className={styles.emptyRow}>
                <Text size="2" color="gray">
                  <FormattedMessage
                    id={
                      isSearch
                        ? "UsersList.empty.search"
                        : "UsersList.empty.normal"
                    }
                  />
                </Text>
              </div>
            ) : null}
          </div>
        </div>
        {!isSearch ? (
          <PaginationWidget
            className={cn(styles.pagination, isEmpty && styles.empty)}
            offset={offset}
            pageSize={pageSize}
            totalCount={totalCount}
            onChangeOffset={onChangeOffset}
          />
        ) : null}
      </div>
      {disableUserDialogData != null ? (
        <AccountStatusDialog
          key={`${disableUserDialogData.accountStatus.id}:${disableUserDialogData.mode}`}
          isHidden={isDisableUserDialogHidden}
          onDismiss={dismissDisableUserDialog}
          accountStatus={disableUserDialogData.accountStatus}
          mode={disableUserDialogData.mode}
        />
      ) : null}
    </>
  );
};

export default UsersList;
