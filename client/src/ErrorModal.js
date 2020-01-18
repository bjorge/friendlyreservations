import React from 'react';
import { Button, Modal, ModalHeader, ModalBody, ModalFooter } from 'reactstrap';


// Example network error:
// {
//   "graphQLErrors": [],
//   "networkError": {},
//   "message": "Network error: Failed to fetch"
// }

// TODO: on network error consider: add button to refresh this page (ex. Redirect or Link)
// example: https://medium.com/@anneeb/redirecting-in-react-4de5e517354a

class ErrorModal extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      modal: true
    };

    this.toggle = this.toggle.bind(this);
  }

  toggle() {
    this.setState({
      modal: !this.state.modal
    });
  }

  render() {
    console.log("modal error page for gql");
    console.log(this.props.error);
    const json = JSON.stringify(this.props.error, null, 2);
    return (
      <div>
        <Modal isOpen={this.state.modal} toggle={this.toggle}>
          <ModalHeader toggle={this.toggle}>Unexpected Error</ModalHeader>
          <ModalBody>
            Error details:<br />
            <pre>{json}</pre>
            Please try again.
          </ModalBody>
          <ModalFooter>
            <Button onClick={this.toggle}>Ok</Button>
          </ModalFooter>
        </Modal>
      </div>
    );
  }
}

export default ErrorModal;


