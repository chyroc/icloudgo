export interface PageText {
  cn: {
    [key: string]: string;
  },
  en: {
    [key: string]: string;
  }
}

const textContent: PageText = {
  en: {
    // login
    loginTitle: 'Login to Admin Account',
    loginAccountPlaceholder: 'Enter Admin Account',
    loginPasswordPlaceholder: 'Enter Admin Password',
    loginButton: 'Login',
    loginingButton: 'Login...',

    // register
    registerTitle: 'Register Admin Account',
    registerButton: 'Register',
    registeringButton: 'Register...',

    // addAccount

    // manageAccount
    email: 'Email',
    totalNumber: 'Total Number',
    downloadedNumber: 'Downloaded Number',
    lastSync: 'Last Sync',
    configure: 'Configure',
    delete: 'Delete',

    // view
    accountManager: 'iCloud Account Manager',
    deleteAccount: 'Delete Account',
    confirmDeleteAccount: 'Are you sure to delete',
    areYes: '?',
    toggleLanguage: 'Switch Language',
    addAccount: 'Add iCloud Account',
    accountPlaceholder: 'Enter iCloud Account',
    passwordPlaceholder: 'Enter iCloud Password',
    twoFACodePlaceholder: 'Enter 2FA Code',
    addButton: 'Add Account',
    addingButton: 'Adding...',

    // config
    icloudPhotoDownloadConfig: 'iCloud Photo Downloader Config',
    iCloudAccount: 'iCloud Account',
    iCloudPassword: 'iCloud Password',
    folderStruct: 'Folder Structure',
    threadNum: 'Thread Num',
    save: 'Save',
    cancel: 'Cancel',
  },
  cn: {
    // login
    loginTitle: '登录管理账号',
    loginAccountPlaceholder: '请输入管理账号',
    loginPasswordPlaceholder: '请输入管理密码',
    loginButton: '登录',
    loginingButton: '登录...',

    // register
    registerTitle: '注册管理账号',
    registerButton: '注册',
    registeringButton: '注册...',

    accountManager: 'iCloud 账号管理',
    deleteAccount: '删除账号',
    confirmDeleteAccount: '确定要删除账号',
    areYes: '吗?',
    toggleLanguage: '切换语言',
    addAccount: '添加 iCloud 账号',
    accountPlaceholder: '请输入 iCloud 账号',
    passwordPlaceholder: '请输入 iCloud 密码',
    twoFACodePlaceholder: '请输入 2FA 代码',
    addButton: '添加账号',
    addingButton: '正在添加...',
    email: '邮箱',
    totalNumber: '总照片数',
    downloadedNumber: '已下载照片数',
    lastSync: '上次同步时间',
    configure: '配置',
    delete: '删除',

    // config
    icloudPhotoDownloadConfig: 'iCloud 照片下载器配置',
    iCloudAccount: 'iCloud 账号',
    iCloudPassword: 'iCloud 密码',
    folderStruct: '文件夹格式',
    threadNum: '并发数',
    save: '保存',
    cancel: '取消',
  }
};

export default textContent