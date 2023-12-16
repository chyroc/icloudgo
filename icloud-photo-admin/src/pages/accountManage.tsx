import React, { useState } from 'react';
import { Button, Modal, ModalBody, ModalContent, ModalFooter, ModalHeader } from '@nextui-org/react';
import textContent from "@/i18n";
import { history } from 'umi'
import { delAccount } from "@/net";

export interface Account {
  id: string;
  email: string;
  totalPhotos: number;
  downloadedPhotos: number;
  lastSync: string;
}


const accountsData: Account[] = [
  // 示例数据，您需要根据实际情况调整
  {id: '1', email: 'example1@icloud.com', totalPhotos: 1000, downloadedPhotos: 800, lastSync: '2021-01-01'},
  {id: '2', email: 'example2@icloud.com', totalPhotos: 300, downloadedPhotos: 40, lastSync: '2022-01-01'},
  // ... 其他账号数据
];

const IndexPage = () => {
  // ... 省略的状态和函数
  const [deleteModal, setDeleteModal] = useState(false)
  const [selectedAccount, setSelectedAccount] = useState<Account | null>(null);
  const [language, setLanguage] = useState('cn');
  const [accountList, setAccountList] = useState<Account[]>(accountsData)

  const handleDeleteClick = async (account: Account) => {
    setSelectedAccount(account);
    setDeleteModal(true)
  };

  const handleDeleteModalChange = async (del: boolean, email: string) => {
    if (del) {
      const {success} = await delAccount(email)
      if (success) {
        setAccountList(accountList.filter(v => v.email != email))
      }
    }
    setDeleteModal(false)
    console.log(del, email)
  }

  const handleConfigClick = (account: Account) => {
    history.push(`/configure/${account.email}`)
  }

  const handleAddAccount = () => {
    history.push(`/addAccount`)
  }

  const toggleLanguage = () => {
    setLanguage(language === 'cn' ? 'en' : 'cn');
  };

  const texts = language === 'cn' ? textContent.cn : textContent.en;


  return (
    <div>
      <div style={{padding: 20, maxWidth: '600px', margin: 'auto'}}>
        <h1 style={{fontSize: '24px', textAlign: 'center'}}>{texts.accountManager}</h1>

        <Button color="primary" style={{textAlign: 'left'}} onClick={handleAddAccount}>
          {texts.addAccount}
        </Button>

        {accountList.map((account) => (
          <div key={account.id}
               style={{border: '1px solid #ccc', padding: '10px', margin: '10px 0', textAlign: 'left'}}>
            <div>{texts.email}: {account.email}</div>
            <div>{texts.totalNumber}: {account.totalPhotos}</div>
            <div>{texts.downloadedNumber}: {account.downloadedPhotos}</div>
            <div>{texts.lastSync}: {account.lastSync}</div>
            <div style={{display: 'flex', justifyContent: 'space-between', marginTop: '10px'}}>
              <Button color="danger" onClick={() => handleDeleteClick(account)}>{texts.delete}</Button>
              <Button onClick={() => handleConfigClick(account)}>{texts.configure}</Button>
            </div>
          </div>
        ))}
      </div>

      <div style={{display: 'flex', justifyContent: 'center', marginTop: '20px'}}>
        <span onClick={toggleLanguage} style={{cursor: 'pointer'}}>
          {language === 'cn' ? '🇨🇳' : '🇺🇸'} {texts.toggleLanguage}
        </span>
      </div>

      {
        !!selectedAccount &&
        <Modal isOpen={deleteModal}>
          <ModalContent>
            {(onClose) => (
              <>
                <ModalHeader className="flex flex-col gap-1">{texts.deleteAccount}</ModalHeader>
                <ModalBody>
                  <p>
                    {texts.confirmDeleteAccount} {selectedAccount.email} {texts.areYes}
                  </p>
                </ModalBody>
                <ModalFooter>
                  <Button color="danger" variant="light"
                          onPress={() => handleDeleteModalChange(true, selectedAccount.email)}>
                    确认删除
                  </Button>
                  <Button variant="light" onPress={() => handleDeleteModalChange(false, selectedAccount.email)}>
                    取消
                  </Button>
                </ModalFooter>
              </>
            )}
          </ModalContent>
        </Modal>
      }
    </div>
  );
};

export default IndexPage;
