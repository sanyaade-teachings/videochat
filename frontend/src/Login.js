import React from 'react';
import TextField from '@material-ui/core/TextField';
import { makeStyles } from '@material-ui/core/styles';
import Button from "@material-ui/core/Button";
import axios from 'axios'
import { useStore } from 'react-redux'
import { restorePreviousUrl } from "./actions";

const useStyles = makeStyles((theme) => ({
  root: {
    '& .MuiTextField-root': {
      margin: theme.spacing(1),
      width: 200,
    },
  },
}));

function Login() {
  const classes = useStyles();

  // EXAMPLE ONLY! Do not do this in a real app.
  // The component will not automatically update if the store state changes
  const store = useStore();

  const [loginDto, setLoginDto] = React.useState({username: "admin", password: "admin"});

  const onLogin = (c) => {
    console.log("on login");

    const params = new URLSearchParams();
    Object.keys(c).forEach((key) => {
      params.append(key, c[key])
    });

    axios.post(`/api/login`, params)
        .then(() => {
            console.log("Store state", store.getState());
            store.dispatch(restorePreviousUrl());
        })
        .catch((error) => {
          // handle error
          console.log("Handling error on login", error.response);
        });
  };

  const handleChangeUsername = event => {
    const dto = {...loginDto, username: event.target.value};
    setLoginDto(dto);
  };

  const handleChangePassword = event => {
    const dto = {...loginDto, password: event.target.value};
    setLoginDto(dto);
  };


  return (
      <form className={classes.root} noValidate autoComplete="off">
        <div>
          <TextField
              id="login"
              label="Login"
              value={loginDto.username}
              onChange={handleChangeUsername}
          />
        </div>
        <div>
          <TextField
              id="password"
              label="Password"
              value={loginDto.password}
              onChange={handleChangePassword}
          />
        </div>
        <Button variant="contained" color="primary"
                onClick={(e) => onLogin(loginDto)}
        >
          Login
        </Button>
      </form>
  );
}

export default (Login);