import { defineMock } from "umi";

export default defineMock({
  'POST /api/login': (req, res) => {
    res.send({
      msg: 'success',
      success: true
    })
  },
  'POST /api/register': (req, res) => {
    res.send({
      msg: 'success',
      success: true,
      duplicate: false,
    })
  },
  'POST /api/addAccount': (req, res) => {
    if (req.body.twoFactorCode === '') {
      res.send({
        msg: 'need needsTwoFactor',
        needsTwoFactor: true
      })
      return
    }

    res.send({
      msg: 'success',
      success: true
    })
  },
  'POST /api/delAccount': (req, res) => {
    res.send({
      msg: 'success',
      success: true
    })
  }
});